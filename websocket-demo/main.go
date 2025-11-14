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
	// 读取查询参数（从 URL 中获取的自定义信息）
	token := c.Query("token")
	userID := c.Query("user_id")
	clientID := c.Query("client_id")

	// 读取请求头（如果客户端通过其他方式设置了请求头）
	authHeader := c.GetHeader("Authorization")
	customHeader := c.GetHeader("X-Custom-Header")

	// 打印连接信息
	log.Printf("New WebSocket connection request:")
	if token != "" {
		log.Printf("  - Token (from query): %s", token)
	}
	if userID != "" {
		log.Printf("  - User ID (from query): %s", userID)
	}
	if clientID != "" {
		log.Printf("  - Client ID (from query): %s", clientID)
	}
	if authHeader != "" {
		log.Printf("  - Authorization header: %s", authHeader)
	}
	if customHeader != "" {
		log.Printf("  - Custom header: %s", customHeader)
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 存储客户端信息（可以扩展为更复杂的客户端管理）
	clientInfo := map[string]interface{}{
		"token":       token,
		"user_id":     userID,
		"client_id":   clientID,
		"auth_header": authHeader,
	}
	log.Printf("Client info: %+v", clientInfo)

	// 注册新客户端
	clients[conn] = true
	log.Printf("New client connected. Total clients: %d", len(clients))

	// 启动一个 goroutine 来处理从客户端接收的消息
	go handleClientMessages(conn, clientInfo)

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
func handleClientMessages(conn *websocket.Conn, clientInfo map[string]interface{}) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message error: %v", err)
			break
		}
		log.Printf("Received from client: %s", message)

		// 尝试解析为 JSON 消息
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			// 处理认证消息
			if msgType, ok := msg["type"].(string); ok && msgType == "auth" {
				log.Printf("Received auth message from client:")
				if token, ok := msg["token"].(string); ok && token != "" {
					clientInfo["token"] = token
					log.Printf("  - Updated token: %s", token)
				}
				if userID, ok := msg["user_id"].(string); ok && userID != "" {
					clientInfo["user_id"] = userID
					log.Printf("  - Updated user_id: %s", userID)
				}
				if clientID, ok := msg["client_id"].(string); ok && clientID != "" {
					clientInfo["client_id"] = clientID
					log.Printf("  - Updated client_id: %s", clientID)
				}
				log.Printf("Updated client info: %+v", clientInfo)
			}
		}
		// 这里可以处理其他类型的客户端消息
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
