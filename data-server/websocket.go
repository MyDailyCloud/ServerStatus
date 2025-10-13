package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketManager WebSocket连接管理器
type WebSocketManager struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	register     chan *Client
	unregister   chan *Client
	clientsMutex sync.RWMutex
}

// Client WebSocket客户端
type Client struct {
	manager    *WebSocketManager
	conn       *websocket.Conn
	send       chan []byte
	projectKey string // 客户端所属的项目密钥
	connected  time.Time
}

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	Type      string      `json:"type"`       // 消息类型: "update", "status", "error"
	Data      interface{} `json:"data"`       // 消息数据
	Timestamp time.Time   `json:"timestamp"`  // 时间戳
	SessionID string      `json:"session_id"` // 服务器Session ID (可选)
	Hostname  string      `json:"hostname"`   // 主机名 (可选)
}

// ServerUpdateMessage 服务器更新消息
type ServerUpdateMessage struct {
	ServerStatus ServerStatus `json:"server_status"`
	Action       string       `json:"action"` // "update", "online", "offline"
}

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 在生产环境中，应该检查Origin
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// NewWebSocketManager 创建新的WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256), // 添加缓冲区避免阻塞
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
	}
}

// Start 启动WebSocket管理器
func (m *WebSocketManager) Start() {
	go m.run()
}

// run 运行WebSocket管理器主循环
func (m *WebSocketManager) run() {
	ticker := time.NewTicker(30 * time.Second) // 心跳检测
	defer ticker.Stop()

	for {
		select {
		case client := <-m.register:
			m.clientsMutex.Lock()
			m.clients[client] = true
			m.clientsMutex.Unlock()
			log.Printf("WebSocket客户端连接: 当前连接数 %d", len(m.clients))

		case client := <-m.unregister:
			m.clientsMutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
				log.Printf("WebSocket客户端断开: 当前连接数 %d", len(m.clients))
			}
			m.clientsMutex.Unlock()

		case message := <-m.broadcast:
			m.clientsMutex.RLock()
			for client := range m.clients {
				select {
				case client.send <- message:
				default:
					// 发送通道已满，关闭客户端连接
					close(client.send)
					delete(m.clients, client)
				}
			}
			m.clientsMutex.RUnlock()

		case <-ticker.C:
			// 发送心跳消息
			m.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送心跳消息
func (m *WebSocketManager) sendHeartbeat() {
	heartbeat := WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"status": "alive"},
	}

	message, err := json.Marshal(heartbeat)
	if err != nil {
		return
	}

	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	for client := range m.clients {
		select {
		case client.send <- message:
		default:
			// 发送失败，标记为需要清理
			go func(c *Client) {
				m.unregister <- c
			}(client)
		}
	}
}

// BroadcastServerUpdate 广播服务器更新消息
func (m *WebSocketManager) BroadcastServerUpdate(serverStatus ServerStatus, action string) {
	update := ServerUpdateMessage{
		ServerStatus: serverStatus,
		Action:       action,
	}

	message := WebSocketMessage{
		Type:      "server_update",
		Data:      update,
		Timestamp: time.Now(),
		SessionID: serverStatus.SessionID,
		Hostname:  serverStatus.Hostname,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("WebSocket消息序列化失败: %v", err)
		return
	}

	select {
	case m.broadcast <- messageBytes:
	default:
		// 广播通道已满，丢弃消息
		log.Println("WebSocket广播通道已满，丢弃消息")
	}
}

// BroadcastToProject 向特定项目组广播消息
func (m *WebSocketManager) BroadcastToProject(projectKey string, message WebSocketMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("WebSocket消息序列化失败: %v", err)
		return
	}

	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	sent := 0
	for client := range m.clients {
		if client.projectKey == projectKey {
			select {
			case client.send <- messageBytes:
				sent++
			default:
				// 发送失败，标记为需要清理
				go func(c *Client) {
					m.unregister <- c
				}(client)
			}
		}
	}

	log.Printf("向项目组 %s 广播消息，发送到 %d 个客户端", projectKey, sent)
}

// GetConnectionCount 获取当前连接数
func (m *WebSocketManager) GetConnectionCount() int {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return len(m.clients)
}

// GetStats 获取WebSocket统计信息
func (m *WebSocketManager) GetStats() map[string]interface{} {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	projectStats := make(map[string]int)
	totalConnections := len(m.clients)

	for client := range m.clients {
		projectStats[client.projectKey]++
	}

	return map[string]interface{}{
		"total_connections": totalConnections,
		"project_stats":     projectStats,
		"timestamp":         time.Now(),
	}
}

// handleWebSocket 处理WebSocket连接请求
func (m *WebSocketManager) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接到WebSocket
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 获取查询参数
	query := r.URL.Query()
	projectKey := query.Get("project_key")
	if projectKey == "" {
		projectKey = "public"
	}

	// 创建客户端
	client := &Client{
		manager:    m,
		conn:       conn,
		send:       make(chan []byte, 256),
		projectKey: projectKey,
		connected:  time.Now(),
	}

	// 注册客户端
	m.register <- client

	// 启动客户端读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 处理WebSocket读取
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		_ = c.conn.Close()
	}()

	// 设置读取超时
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// 读取消息
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误 [%s]: %v", c.projectKey, err)
			}
			break
		}

		// 处理客户端消息（这里可以实现客户端到服务器的通信）
		if len(message) > 0 {
			log.Printf("收到WebSocket消息 [%s]: %s", c.projectKey, string(message))
		}
	}
}

// writePump 处理WebSocket写入
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second) // ping间隔
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 发送通道已关闭
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		_, _ = w.Write(message)

		// 添加队列中的其他消息
		n := len(c.send)
		for i := 0; i < n; i++ {
			_, _ = w.Write([]byte{'\n'})
			_, _ = w.Write(<-c.send)
		}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// 发送ping
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 全局WebSocket管理器实例
var webSocketManager = NewWebSocketManager()
