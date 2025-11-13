package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// 存储所有活跃的 WebSocket 连接
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

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
	clients[conn] = true
	log.Printf("New client connected. Total clients: %d", len(clients))

	// 启动一个 goroutine 来处理从客户端接收的消息
	go handleClientMessages(conn)

	// 保持连接活跃，等待客户端断开
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			delete(clients, conn)
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
		message := <-broadcast
		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		// 向所有客户端发送消息
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, messageJSON)
			if err != nil {
				log.Printf("Write message error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// startTimer 启动定时推送任务
func startTimer() {
	ticker := time.NewTicker(5 * time.Second) // 每 5 秒推送一次
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
		broadcast <- message
		log.Printf("Broadcasted message ID: %d to %d clients", messageID, len(clients))
	}
}

func main() {
	// 启动广播 goroutine
	go broadcastMessage()

	// 启动定时推送 goroutine
	go startTimer()

	// 创建 Gin 路由
	r := gin.Default()

	// 静态文件服务（用于提供测试页面）
	r.Static("/static", "./static")

	// WebSocket 端点
	r.GET("/ws", handleWebSocket)

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"clients": len(clients),
		})
	})

	// 根路径，返回 HTML 测试页面
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	log.Println("WebSocket server starting on :8080")
	log.Println("WebSocket endpoint: ws://localhost:8080/ws")
	log.Println("Test page: http://localhost:8080/")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
