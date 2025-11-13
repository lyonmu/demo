package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	capi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
)

var (
	ConsulClient *capi.Client
	Router       *gin.Engine
	// 存储所有活跃的 WebSocket 连接
	Clients   = make(map[*websocket.Conn]bool)
	Broadcast = make(chan Message)
)

// Message 定义推送的消息结构体
type Message struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Data      DataInfo  `json:"data"`
}

// DataInfo 消息中的额外数据
type DataInfo struct {
	Status string  `json:"status"`
	Value  float64 `json:"value"`
	Count  int     `json:"count"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源的连接，生产环境应该检查具体的来源
		return true
	},
}

// handleWebSocket 处理 WebSocket 连接
func handleWebSocket(c *gin.Context) {
	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 注册新客户端
	Clients[conn] = true
	log.Printf("New client connected. Total clients: %d", len(Clients))

	// 启动一个 goroutine 来处理从客户端接收的消息
	go handleClientMessages(conn)

	// 保持连接活跃，等待客户端断开
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			delete(Clients, conn)
			break
		}
	}
}

// handleClientMessages 处理来自客户端的消息
func handleClientMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message error: %v", err)
			break
		}
		log.Printf("Received from client: %s", message)
		// 这里可以处理客户端发送的消息
	}
}

// broadcastMessage 向所有客户端广播消息
func broadcastMessage() {
	for {
		message := <-Broadcast
		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		// 向所有客户端发送消息
		for client := range Clients {
			err := client.WriteMessage(websocket.TextMessage, messageJSON)
			if err != nil {
				log.Printf("Write message error: %v", err)
				client.Close()
				delete(Clients, client)
			}
		}
	}
}

// startTimer 启动定时推送任务
func startTimer() {
	ticker := time.NewTicker(1 * time.Second) // 每 1 秒推送一次
	defer ticker.Stop()

	messageID := 0
	for range ticker.C {
		messageID++
		message := Message{
			ID:        messageID,
			Type:      "notification",
			Content:   "这是一条定时推送的消息",
			Timestamp: time.Now(),
			Data: DataInfo{
				Status: "active",
				Value:  float64(messageID) * 1.5,
				Count:  messageID * 10,
			},
		}
		Broadcast <- message
		log.Printf("Broadcasted message ID: %d to %d clients", messageID, len(Clients))
	}
}

func initConsul() error {

	config := capi.DefaultConfig()
	config.Address = "192.168.100.156:8500"
	config.Token = "0q8Rz0ElbVb2uRsTc4Cj3LTAnwwX4FhmSFAvblwToMw"
	client, err := capi.NewClient(config)
	if err != nil {
		return err
	}
	ConsulClient = client
	return nil

}

func RegisterMetrics(engine *gin.Engine) error {
	reg := prometheus.NewRegistry()
	collectorsList := []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	}
	for _, v := range collectorsList {
		if err := reg.Register(v); err != nil {
			return err
		}
	}

	engine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling:     promhttp.ContinueOnError,
		EnableOpenMetrics: true,
		Registry:          reg,
	})))
	return nil
}

func initGin() {
	// Set mode
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.ErrorLogger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	// Router group
	RouterGroup := router.Group("demo")
	RouterGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	RouterGroup.GET("/ws", handleWebSocket)
	if err := RegisterMetrics(router); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register metrics: %v", err)
		os.Exit(1)
	}
	Router = router

}

func main() {

	if err := initConsul(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Consul: %v", err)
		os.Exit(1)
	}

	initGin()
	go broadcastMessage()
	go startTimer()

	// Main listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v", err)
		os.Exit(1)
	}

	// Create a cmux.
	m := cmux.New(listener)

	// http://127.0.0.1:20416/demo/health
	regEnvoy := &capi.AgentServiceRegistration{
		Name:    "demo",
		Port:    8080,
		Address: "192.168.8.24",
		ID:      fmt.Sprintf("%s:%d", "192.168.8.24", 8080),
		Check: &capi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/demo/health", "192.168.8.24", 8080),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "10s",
		},
		Tags: []string{"demo", "sentinel"},
		Meta: map[string]string{
			"version":       "1.0.0",
			"start_time":    time.Now().Format(time.DateTime),
			"router_prefix": "demo",
			"no_auth":       "false",
		},
	}

	ConsulClient.Agent().ServiceRegister(regEnvoy)

	httpL := m.Match(cmux.HTTP1Fast())

	// Create HTTP server with gin router
	httpServer := &http.Server{
		Handler: Router,
	}

	// Run HTTP server in a goroutine to avoid blocking
	go func() {
		if err := httpServer.Serve(httpL); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Failed to serve HTTP: %v", err)
		}
	}()

	// Start serving! This will block and handle all connections
	if err := m.Serve(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		fmt.Fprintf(os.Stderr, "Failed to serve: %v", err)
		os.Exit(1)
	}

}
