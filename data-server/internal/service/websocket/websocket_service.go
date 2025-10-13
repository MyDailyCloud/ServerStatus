package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeError        MessageType = "error"
	MessageTypeSystem       MessageType = "system"
	MessageTypeServerUpdate MessageType = "server_update"
	MessageTypeServerStatus MessageType = "server_status"
	MessageTypeAlert        MessageType = "alert"
	MessageTypeHeartbeat    MessageType = "heartbeat"
	MessageTypeAuth         MessageType = "auth"
	MessageTypeSubscribe    MessageType = "subscribe"
	MessageTypeUnsubscribe  MessageType = "unsubscribe"
)

// ClientInfo 客户端信息
type ClientInfo struct {
	ID          string                 `json:"id"`
	Conn        *websocket.Conn        `json:"-"`
	ProjectKey  string                 `json:"project_key"`
	Permissions []string               `json:"permissions"`
	UserAgent   string                 `json:"user_agent"`
	RemoteAddr  string                 `json:"remote_addr"`
	ConnectedAt time.Time              `json:"connected_at"`
	LastSeen    time.Time              `json:"last_seen"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type      MessageType            `json:"type"`
	ID        string                 `json:"id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ServerData 服务器数据消息
type ServerData struct {
	SessionID string                 `json:"session_id"`
	Hostname  string                 `json:"hostname"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// AlertData 告警数据
type AlertData struct {
	ServerID  string                 `json:"server_id"`
	Hostname  string                 `json:"hostname"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	Events   []MessageType `json:"events"`
	Servers  []string      `json:"servers,omitempty"`
	Projects []string      `json:"projects,omitempty"`
}

// AuthRequest 认证请求
type AuthRequest struct {
	AccessKey string `json:"access_key"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Success     bool     `json:"success"`
	ProjectKey  string   `json:"project_key,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Message     string   `json:"message,omitempty"`
}

// Stats 统计信息
type Stats struct {
	TotalConnections   int            `json:"total_connections"`
	ActiveConnections  int            `json:"active_connections"`
	ProjectConnections map[string]int `json:"project_connections"`
	MessagesSent       int64          `json:"messages_sent"`
	MessagesReceived   int64          `json:"messages_received"`
	AverageMessageSize int64          `json:"average_message_size"`
	LastError          string         `json:"last_error,omitempty"`
	LastActivity       time.Time      `json:"last_activity"`
}

// Config WebSocket服务配置
type Config struct {
	// 连接配置
	ReadTimeout    time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout" json:"write_timeout"`
	PingPeriod     time.Duration `yaml:"ping_period" json:"ping_period"`
	PongWait       time.Duration `yaml:"pong_wait" json:"pong_wait"`
	MaxMessageSize int64         `yaml:"max_message_size" json:"max_message_size"`

	// 连接限制
	MaxConnectionsPerProject int `yaml:"max_connections_per_project" json:"max_connections_per_project"`
	MaxTotalConnections      int `yaml:"max_total_connections" json:"max_total_connections"`

	// 认证配置
	AuthTimeout time.Duration `yaml:"auth_timeout" json:"auth_timeout"`
	RequireAuth bool          `yaml:"require_auth" json:"require_auth"`

	// 统计配置
	StatsInterval time.Duration `yaml:"stats_interval" json:"stats_interval"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		ReadTimeout:              60 * time.Second,
		WriteTimeout:             10 * time.Second,
		PingPeriod:               54 * time.Second,
		PongWait:                 60 * time.Second,
		MaxMessageSize:           1024 * 1024, // 1MB
		MaxConnectionsPerProject: 100,
		MaxTotalConnections:      1000,
		AuthTimeout:              10 * time.Second,
		RequireAuth:              true,
		StatsInterval:            30 * time.Second,
	}
}

// WebSocketService WebSocket服务
type WebSocketService struct {
	clients       map[string]*ClientInfo
	clientsMutex  sync.RWMutex
	subscriptions map[string]map[MessageType]bool
	subsMutex     sync.RWMutex
	stats         *Stats
	statsMutex    sync.RWMutex

	// 依赖
	serverRepo    repository.ServerRepository
	cacheRepo     repository.CacheRepository
	accessKeyRepo repository.AccessKeyRepository
	authService   AuthService
	logger        logger.Logger

	// 配置和状态
	config   *Config
	upgrader websocket.Upgrader
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// AuthService 认证服务接口
type AuthService interface {
	ValidateAccessKey(ctx context.Context, req *AuthRequest) (*AuthResult, error)
}

// AuthResult 认证结果
type AuthResult struct {
	Success     bool          `json:"success"`
	ProjectKey  string        `json:"project_key,omitempty"`
	Permissions []string      `json:"permissions,omitempty"`
	Message     string        `json:"message,omitempty"`
	TTL         time.Duration `json:"ttl,omitempty"`
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService(
	serverRepo repository.ServerRepository,
	cacheRepo repository.CacheRepository,
	accessKeyRepo repository.AccessKeyRepository,
	authService AuthService,
	logger logger.Logger,
	config *Config,
) *WebSocketService {
	if config == nil {
		config = DefaultConfig()
	}

	service := &WebSocketService{
		clients:       make(map[string]*ClientInfo),
		subscriptions: make(map[string]map[MessageType]bool),
		stats: &Stats{
			ProjectConnections: make(map[string]int),
			LastActivity:       time.Now(),
		},
		serverRepo:    serverRepo,
		cacheRepo:     cacheRepo,
		accessKeyRepo: accessKeyRepo,
		authService:   authService,
		logger:        logger,
		config:        config,
		shutdown:      make(chan struct{}),
	}

	// 配置WebSocket upgrader
	service.upgrader = websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			// 在生产环境中应该检查Origin
			return true
		},
	}

	// 启动统计协程
	service.wg.Add(1)
	go service.statsWorker()

	return service
}

// HandleConnection 处理WebSocket连接
func (s *WebSocketService) HandleConnection(ctx context.Context, conn *websocket.Conn, remoteAddr, userAgent string) (*ClientInfo, error) {
	// 检查连接限制
	if err := s.checkConnectionLimits(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("connection limit exceeded: %w", err)
	}

	// 生成客户端ID
	clientID := s.generateClientID()

	// 创建客户端信息
	client := &ClientInfo{
		ID:          clientID,
		Conn:        conn,
		UserAgent:   userAgent,
		RemoteAddr:  remoteAddr,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// 注册客户端
	s.registerClient(client)

	s.logger.WithFields(map[string]interface{}{
		"client_id":   clientID,
		"remote_addr": remoteAddr,
		"user_agent":  userAgent,
	}).Info("WebSocket client connected")

	// 设置连接参数
	conn.SetReadLimit(s.config.MaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		client.LastSeen = time.Now()
		return nil
	})

	// 启动读写协程
	s.wg.Add(2)
	go s.readPump(ctx, client)
	go s.writePump(ctx, client)

	// 如果需要认证，发送认证请求
	if s.config.RequireAuth {
		_ = s.sendMessage(client, &WebSocketMessage{
			Type:      MessageTypeAuth,
			Timestamp: time.Now(),
			Data:      map[string]string{"message": "authentication required"},
		})
	}

	return client, nil
}

// readPump 读取消息协程
func (s *WebSocketService) readPump(ctx context.Context, client *ClientInfo) {
	defer s.wg.Done()
	defer s.unregisterClient(client)

	conn := client.Conn
	defer conn.Close()

	// 设置读取超时
	_ = conn.SetReadDeadline(time.Now().Add(s.config.AuthTimeout))
	authTimer := time.NewTimer(s.config.AuthTimeout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		case <-authTimer.C:
			if s.config.RequireAuth && client.ProjectKey == "" {
				s.sendError(client, "authentication timeout")
				return
			}
		}

		// 读取消息
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithField("error", err).Warn("WebSocket read error")
			}
			break
		}

		// 处理消息
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			s.handleMessage(client, data)

		// 停止认证计时器
		if authTimer.Stop() {
			// 设置正常的读取超时
			_ = conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		}
		}

		// 更新最后活跃时间
		client.LastSeen = time.Now()
		s.updateLastActivity()
	}
}

// writePump 写入消息协程
func (s *WebSocketService) writePump(ctx context.Context, client *ClientInfo) {
	defer s.wg.Done()
	defer client.Conn.Close()

	conn := client.Conn
	ticker := time.NewTicker(s.config.PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		case <-ticker.C:
			// 发送ping
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.WithField("error", err).Warn("WebSocket ping failed")
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (s *WebSocketService) handleMessage(client *ClientInfo, data []byte) {
	var message WebSocketMessage
	if err := json.Unmarshal(data, &message); err != nil {
		s.sendError(client, "invalid message format")
		return
	}

	// 更新统计
	s.incrementMessagesReceived()

	switch message.Type {
	case MessageTypeAuth:
		s.handleAuthMessage(client, message)
	case MessageTypeSubscribe:
		s.handleSubscribeMessage(client, message)
	case MessageTypeUnsubscribe:
		s.handleUnsubscribeMessage(client, message)
	case MessageTypeHeartbeat:
		s.handleHeartbeatMessage(client, message)
	default:
		s.sendError(client, fmt.Sprintf("unsupported message type: %s", message.Type))
	}
}

// handleAuthMessage 处理认证消息
func (s *WebSocketService) handleAuthMessage(client *ClientInfo, message WebSocketMessage) {
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		s.sendError(client, "invalid auth data")
		return
	}

	var authReq AuthRequest
	if err := json.Unmarshal(dataBytes, &authReq); err != nil {
		s.sendError(client, "invalid auth request")
		return
	}

	// 验证访问密钥
	ctx := context.Background()
	result, err := s.authService.ValidateAccessKey(ctx, &authReq)
	if err != nil {
		s.sendError(client, "authentication failed")
		return
	}

	if result.Success {
		// 更新客户端信息
		client.ProjectKey = result.ProjectKey
		client.Permissions = result.Permissions
		client.LastSeen = time.Now()

		// 更新项目连接统计
		s.updateProjectConnections(result.ProjectKey, 1)

		// 发送认证成功响应
		_ = s.sendMessage(client, &WebSocketMessage{
			Type:      MessageTypeAuth,
			Timestamp: time.Now(),
			Data: AuthResponse{
				Success:     true,
				ProjectKey:  result.ProjectKey,
				Permissions: result.Permissions,
				Message:     "authentication successful",
			},
		})

		s.logger.WithFields(map[string]interface{}{
			"client_id":   client.ID,
			"project_key": result.ProjectKey,
		}).Info("Client authenticated")
	} else {
		_ = s.sendMessage(client, &WebSocketMessage{
			Type:      MessageTypeAuth,
			Timestamp: time.Now(),
			Data: AuthResponse{
				Success: false,
				Message: result.Message,
			},
		})
	}
}

// handleSubscribeMessage 处理订阅消息
func (s *WebSocketService) handleSubscribeMessage(client *ClientInfo, message WebSocketMessage) {
	if client.ProjectKey == "" && s.config.RequireAuth {
		s.sendError(client, "authentication required for subscription")
		return
	}

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		s.sendError(client, "invalid subscription data")
		return
	}

	var subscribeReq SubscribeRequest
	if err := json.Unmarshal(dataBytes, &subscribeReq); err != nil {
		s.sendError(client, "invalid subscription request")
		return
	}

	// 验证权限
	if !s.hasPermission(client, "websocket") {
		s.sendError(client, "insufficient permissions for websocket subscription")
		return
	}

	// 添加订阅
	s.subsMutex.Lock()
	if s.subscriptions[client.ID] == nil {
		s.subscriptions[client.ID] = make(map[MessageType]bool)
	}

	for _, eventType := range subscribeReq.Events {
		s.subscriptions[client.ID][eventType] = true
	}
	s.subsMutex.Unlock()

	// 发送订阅成功响应
	_ = s.sendMessage(client, &WebSocketMessage{
		Type:      MessageTypeSystem,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": "subscription successful",
			"events":  subscribeReq.Events,
		},
	})

	s.logger.WithFields(map[string]interface{}{
		"client_id": client.ID,
		"events":    subscribeReq.Events,
	}).Info("Client subscribed to events")
}

// handleUnsubscribeMessage 处理取消订阅消息
func (s *WebSocketService) handleUnsubscribeMessage(client *ClientInfo, message WebSocketMessage) {
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		s.sendError(client, "invalid unsubscribe data")
		return
	}

	var unsubscribeReq SubscribeRequest
	if err := json.Unmarshal(dataBytes, &unsubscribeReq); err != nil {
		s.sendError(client, "invalid unsubscribe request")
		return
	}

	// 移除订阅
	s.subsMutex.Lock()
	if s.subscriptions[client.ID] != nil {
		for _, eventType := range unsubscribeReq.Events {
			delete(s.subscriptions[client.ID], eventType)
		}
	}
	s.subsMutex.Unlock()

	// 发送取消订阅成功响应
	_ = s.sendMessage(client, &WebSocketMessage{
		Type:      MessageTypeSystem,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": "unsubscription successful",
			"events":  unsubscribeReq.Events,
		},
	})
}

// handleHeartbeatMessage 处理心跳消息
func (s *WebSocketService) handleHeartbeatMessage(client *ClientInfo, message WebSocketMessage) {
	client.LastSeen = time.Now()

	// 发送心跳响应
	_ = s.sendMessage(client, &WebSocketMessage{
		Type:      MessageTypeHeartbeat,
		Timestamp: time.Now(),
		Data:      map[string]string{"status": "ok"},
	})
}

// BroadcastServerUpdate 广播服务器更新
func (s *WebSocketService) BroadcastServerUpdate(ctx context.Context, serverData *models.SystemInfo) error {
	serverUpdate := &ServerData{
		SessionID: serverData.SessionID,
		Hostname:  serverData.Hostname,
		Status:    "online",
		Data: map[string]interface{}{
			"cpu":            serverData.CPU,
			"memory":         serverData.Memory,
			"disk":           serverData.Disk,
			"network":        serverData.Network,
			"gpu":            serverData.GPU,
			"gpus":           serverData.GPUs,
			"os":             serverData.OS,
			"temperature":    serverData.Temperature,
			"user_resources": serverData.UserResources,
		},
		UpdatedAt: serverData.Timestamp,
	}

	return s.broadcastMessage(MessageTypeServerUpdate, serverUpdate, serverData.ProjectKey)
}

// BroadcastAlert 广播告警
func (s *WebSocketService) BroadcastAlert(ctx context.Context, alert *AlertData) error {
	return s.broadcastMessage(MessageTypeAlert, alert, "")
}

// broadcastMessage 广播消息
func (s *WebSocketService) broadcastMessage(messageType MessageType, data interface{}, projectKey string) error {
	message := &WebSocketMessage{
		Type:      messageType,
		Timestamp: time.Now(),
		Data:      data,
	}

	s.clientsMutex.RLock()
	s.subsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	defer s.subsMutex.RUnlock()

	sentCount := 0
	for clientID, client := range s.clients {
		// 检查项目权限
		if projectKey != "" && client.ProjectKey != projectKey {
			continue
		}

		// 检查订阅
		if subscriptions, ok := s.subscriptions[clientID]; ok {
			if subscriptions[messageType] || subscriptions[MessageTypeServerUpdate] {
				if err := s.sendMessage(client, message); err != nil {
					s.logger.WithFields(map[string]interface{}{
						"client_id": clientID,
						"error":     err,
					}).Warn("Failed to send message to client")
				} else {
					sentCount++
				}
			}
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"message_type": messageType,
		"project_key":  projectKey,
		"sent_count":   sentCount,
	}).Debug("Broadcast message sent")

	return nil
}

// sendMessage 发送消息给客户端
func (s *WebSocketService) sendMessage(client *ClientInfo, message *WebSocketMessage) error {
	conn := client.Conn
	if conn == nil {
		return fmt.Errorf("client connection is nil")
	}

	// 设置写入超时
	_ = conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

	// 序列化消息
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发送消息
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// 更新统计
	s.incrementMessagesSent(int64(len(data)))

	return nil
}

// sendError 发送错误消息
func (s *WebSocketService) sendError(client *ClientInfo, errorMsg string) {
	message := &WebSocketMessage{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Data:      map[string]string{"error": errorMsg},
	}
	_ = s.sendMessage(client, message)
}

// registerClient 注册客户端
func (s *WebSocketService) registerClient(client *ClientInfo) {
	s.clientsMutex.Lock()
	s.clients[client.ID] = client
	s.clientsMutex.Unlock()

	// 更新统计
	s.statsMutex.Lock()
	s.stats.TotalConnections++
	s.stats.ActiveConnections++
	s.statsMutex.Unlock()
}

// unregisterClient 注销客户端
func (s *WebSocketService) unregisterClient(client *ClientInfo) {
	s.clientsMutex.Lock()
	if _, exists := s.clients[client.ID]; exists {
		delete(s.clients, client.ID)
	}
	s.clientsMutex.Unlock()

	// 移除订阅
	s.subsMutex.Lock()
	delete(s.subscriptions, client.ID)
	s.subsMutex.Unlock()

	// 更新项目连接统计
	if client.ProjectKey != "" {
		s.updateProjectConnections(client.ProjectKey, -1)
	}

	// 更新统计
	s.statsMutex.Lock()
	s.stats.ActiveConnections--
	s.statsMutex.Unlock()

	s.logger.WithField("client_id", client.ID).Info("Client disconnected")
}

// checkConnectionLimits 检查连接限制
func (s *WebSocketService) checkConnectionLimits() error {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	if s.config.MaxTotalConnections > 0 && s.stats.ActiveConnections >= s.config.MaxTotalConnections {
		return fmt.Errorf("maximum total connections (%d) reached", s.config.MaxTotalConnections)
	}

	return nil
}

// hasPermission 检查权限
func (s *WebSocketService) hasPermission(client *ClientInfo, permission string) bool {
	if len(client.Permissions) == 0 {
		return true // 如果没有权限列表，允许所有操作
	}

	for _, p := range client.Permissions {
		if p == permission || p == "admin" {
			return true
		}
	}
	return false
}

// generateClientID 生成客户端ID
func (s *WebSocketService) generateClientID() string {
	return fmt.Sprintf("client_%d_%d", time.Now().UnixNano(), len(s.clients))
}

// updateProjectConnections 更新项目连接统计
func (s *WebSocketService) updateProjectConnections(projectKey string, delta int) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.ProjectConnections[projectKey] += delta
	if s.stats.ProjectConnections[projectKey] <= 0 {
		delete(s.stats.ProjectConnections, projectKey)
	}
}

// updateLastActivity 更新最后活跃时间
func (s *WebSocketService) updateLastActivity() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.LastActivity = time.Now()
}

// incrementMessagesSent 增加发送消息计数
func (s *WebSocketService) incrementMessagesSent(size int64) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.MessagesSent++

	// 更新平均消息大小
	total := s.stats.MessagesSent
	if total > 0 {
		s.stats.AverageMessageSize = (s.stats.AverageMessageSize*(total-1) + size) / total
	}
}

// incrementMessagesReceived 增加接收消息计数
func (s *WebSocketService) incrementMessagesReceived() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.MessagesReceived++
}

// statsWorker 统计协程
func (s *WebSocketService) statsWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown:
			return
		case <-ticker.C:
			s.performStatsUpdate()
		}
	}
}

// performStatsUpdate 执行统计更新
func (s *WebSocketService) performStatsUpdate() {
	// 清理断开的连接
	s.cleanupDisconnectedClients()

	// 记录统计信息
	s.statsMutex.RLock()
	stats := *s.stats
	s.statsMutex.RUnlock()

	s.logger.WithFields(map[string]interface{}{
		"active_connections":   stats.ActiveConnections,
		"total_connections":    stats.TotalConnections,
		"messages_sent":        stats.MessagesSent,
		"messages_received":    stats.MessagesReceived,
		"average_message_size": stats.AverageMessageSize,
		"project_connections":  len(stats.ProjectConnections),
	}).Debug("WebSocket service statistics")
}

// cleanupDisconnectedClients 清理断开的连接
func (s *WebSocketService) cleanupDisconnectedClients() {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for clientID, client := range s.clients {
		// 检查连接是否活跃
		if time.Since(client.LastSeen) > s.config.PongWait*2 {
			s.logger.WithField("client_id", clientID).Warn("Cleaning up inactive client")
			_ = client.Conn.Close()
			delete(s.clients, clientID)

			// 更新统计
			s.statsMutex.Lock()
			s.stats.ActiveConnections--
			s.statsMutex.Unlock()
		}
	}
}

// GetStats 获取统计信息
func (s *WebSocketService) GetStats() *Stats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	stats := *s.stats
	return &stats
}

// GetClientCount 获取客户端数量
func (s *WebSocketService) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}

// GetProjectClientCount 获取项目客户端数量
func (s *WebSocketService) GetProjectClientCount(projectKey string) int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	count := 0
	for _, client := range s.clients {
		if client.ProjectKey == projectKey {
			count++
		}
	}
	return count
}

// GetConnectedClients 获取连接的客户端列表
func (s *WebSocketService) GetConnectedClients() []*ClientInfo {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	clients := make([]*ClientInfo, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

// Shutdown 关闭服务
func (s *WebSocketService) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down WebSocket service")

	// 关闭所有连接
	s.clientsMutex.Lock()
	for _, client := range s.clients {
		_ = client.Conn.Close()
	}
	s.clientsMutex.Unlock()

	// 发送关闭信号
	close(s.shutdown)

	// 等待所有协程完成
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("WebSocket service shutdown completed")
		return nil
	case <-ctx.Done():
		s.logger.Warn("WebSocket service shutdown timeout")
		return ctx.Err()
	}
}

// GetConfig 获取配置
func (s *WebSocketService) GetConfig() *Config {
	return s.config
}

// UpdateConfig 更新配置
func (s *WebSocketService) UpdateConfig(config *Config) {
	if config != nil {
		s.config = config
		s.logger.WithField("config", "updated").Info("WebSocket service configuration updated")
	}
}
