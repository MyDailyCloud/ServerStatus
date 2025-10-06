package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Mock repositories for testing
type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) CreateServer(ctx context.Context, server *models.ServerInfo) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepository) GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.ServerInfo), args.Error(1)
}

func (m *MockServerRepository) UpdateServer(ctx context.Context, server *models.ServerInfo) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepository) DeleteServer(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockServerRepository) GetAllServers(ctx context.Context, projectKey string, offset, limit int) ([]*models.ServerInfo, error) {
	args := m.Called(ctx, projectKey, offset, limit)
	return args.Get(0).([]*models.ServerInfo), args.Error(1)
}

func (m *MockServerRepository) GetServersByHostname(ctx context.Context, hostname string) ([]*models.ServerInfo, error) {
	args := m.Called(ctx, hostname)
	return args.Get(0).([]*models.ServerInfo), args.Error(1)
}

func (m *MockServerRepository) GetServerCount(ctx context.Context, projectKey string) (int, error) {
	args := m.Called(ctx, projectKey)
	return args.Int(0), args.Error(1)
}

func (m *MockServerRepository) UpdateLastSeen(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockServerRepository) GetOnlineServers(ctx context.Context, projectKey string, timeout time.Duration) ([]*models.ServerInfo, error) {
	args := m.Called(ctx, projectKey, timeout)
	return args.Get(0).([]*models.ServerInfo), args.Error(1)
}

func (m *MockServerRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockServerRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheRepository) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	args := m.Called(ctx, items, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCacheRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCacheRepository) ClearPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	args := m.Called(ctx, pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCacheRepository) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCacheRepository) GetType() string {
	args := m.Called()
	return args.String(0)
}

type MockAccessKeyRepository struct {
	mock.Mock
}

func (m *MockAccessKeyRepository) SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error {
	args := m.Called(ctx, cacheKey, accessKey)
	return args.Error(0)
}

func (m *MockAccessKeyRepository) GetAccessKey(ctx context.Context, accessKey string) (string, error) {
	args := m.Called(ctx, accessKey)
	return args.String(0), args.Error(1)
}

func (m *MockAccessKeyRepository) DeleteAccessKey(ctx context.Context, accessKey string) error {
	args := m.Called(ctx, accessKey)
	return args.Error(0)
}

func (m *MockAccessKeyRepository) GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error) {
	args := m.Called(ctx, serverKey, projectKey)
	return args.String(0), args.Error(1)
}

func (m *MockAccessKeyRepository) ValidateAccessKey(ctx context.Context, accessKey string) (string, error) {
	args := m.Called(ctx, accessKey)
	return args.String(0), args.Error(1)
}

func (m *MockAccessKeyRepository) CleanupExpiredKeys(ctx context.Context, ttl time.Duration) error {
	args := m.Called(ctx, ttl)
	return args.Error(0)
}

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateAccessKey(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*AuthResult), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Panic(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Panicf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	return args.Get(0).(logger.Logger)
}

// Test WebSocket service creation
func TestNewWebSocketService(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	// Create service with default config
	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		nil,
	)

	require.NotNil(t, service)
	assert.Equal(t, DefaultConfig(), service.config)
	assert.NotNil(t, service.clients)
	assert.NotNil(t, service.subscriptions)
	assert.NotNil(t, service.stats)

	// Verify dependencies
	assert.Equal(t, mockServerRepo, service.serverRepo)
	assert.Equal(t, mockCacheRepo, service.cacheRepo)
	assert.Equal(t, mockAccessKeyRepo, service.accessKeyRepo)
	assert.Equal(t, mockAuthService, service.authService)
	assert.Equal(t, mockLogger, service.logger)
}

func TestNewWebSocketService_CustomConfig(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	// Create custom config
	customConfig := &Config{
		ReadTimeout:              30 * time.Second,
		WriteTimeout:             5 * time.Second,
		PingPeriod:               25 * time.Second,
		MaxTotalConnections:      500,
		RequireAuth:              false,
		StatsInterval:            15 * time.Second,
	}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		customConfig,
	)

	require.NotNil(t, service)
	assert.Equal(t, customConfig, service.config)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.ReadTimeout > 0)
	assert.True(t, config.WriteTimeout > 0)
	assert.True(t, config.PingPeriod > 0)
	assert.True(t, config.PongWait > 0)
	assert.True(t, config.MaxMessageSize > 0)
	assert.True(t, config.MaxConnectionsPerProject > 0)
	assert.True(t, config.MaxTotalConnections > 0)
	assert.True(t, config.AuthTimeout > 0)
	assert.True(t, config.StatsInterval > 0)
}

func TestWebSocketService_GetStats(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Get initial stats
	stats := service.GetStats()
	require.NotNil(t, stats)
	assert.Equal(t, 0, stats.ActiveConnections)
	assert.Equal(t, 0, stats.TotalConnections)
	assert.Equal(t, int64(0), stats.MessagesSent)
	assert.Equal(t, int64(0), stats.MessagesReceived)
	assert.Equal(t, int64(0), stats.AverageMessageSize)
	assert.NotNil(t, stats.ProjectConnections)
}

func TestWebSocketService_GetClientCount(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Initially no clients
	assert.Equal(t, 0, service.GetClientCount())
}

func TestWebSocketService_GetProjectClientCount(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Initially no clients for any project
	assert.Equal(t, 0, service.GetProjectClientCount("test-project"))
}

func TestWebSocketService_BroadcastServerUpdate(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Setup logger mock expectations
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Create test server data
	serverData := &models.SystemInfo{
		SessionID:  "test-session",
		Hostname:   "test-host",
		ProjectKey: "test-project",
		Timestamp:  time.Now(),
		CPU: models.CPUInfo{
			UsagePercent: 75.5,
		},
		Memory: models.MemInfo{
			Total: 1024,
			Used:  512,
		},
		Disk: models.DiskInfo{
			Total: 2048,
			Used:  1024,
		},
		Network: models.NetInfo{
			BytesSent: 1000,
			BytesRecv: 2000,
		},
	}

	// Broadcast without clients should not error
	ctx := context.Background()
	err := service.BroadcastServerUpdate(ctx, serverData)
	assert.NoError(t, err)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestWebSocketService_BroadcastAlert(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Setup logger mock expectations
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Create test alert
	alert := &AlertData{
		ServerID:  "test-server",
		Hostname:  "test-host",
		Level:     "warning",
		Message:   "High CPU usage",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"cpu": 95.5},
	}

	// Broadcast without clients should not error
	ctx := context.Background()
	err := service.BroadcastAlert(ctx, alert)
	assert.NoError(t, err)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestWebSocketService_Configuration(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Setup logger mock expectations
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Test get config
	config := service.GetConfig()
	require.NotNil(t, config)
	assert.Equal(t, DefaultConfig().ReadTimeout, config.ReadTimeout)

	// Test update config
	newConfig := &Config{
		ReadTimeout:         100 * time.Second,
		MaxTotalConnections: 2000,
	}
	service.UpdateConfig(newConfig)

	updatedConfig := service.GetConfig()
	assert.Equal(t, 100*time.Second, updatedConfig.ReadTimeout)
	assert.Equal(t, 2000, updatedConfig.MaxTotalConnections)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestWebSocketService_Shutdown(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	// Setup logger mock expectations for shutdown
	mockLogger.On("Info", mock.Anything).Return()

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		DefaultConfig(),
	)

	// Shutdown should complete without error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := service.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestFactory_CreateService(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create factory
	factory := NewFactory(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger)
	require.NotNil(t, factory)

	// Test create service with custom config
	customConfig := &Config{
		MaxTotalConnections: 500,
		RequireAuth:         false,
	}
	service := factory.CreateService(customConfig)
	require.NotNil(t, service)
	assert.Equal(t, 500, service.config.MaxTotalConnections)
	assert.False(t, service.config.RequireAuth)

	// Test create default service
	defaultService := factory.CreateDefaultService()
	require.NotNil(t, defaultService)
	assert.Equal(t, DefaultConfig().MaxTotalConnections, defaultService.config.MaxTotalConnections)
}

func TestWebSocketMessage_JSONSerialization(t *testing.T) {
	// Test message serialization
	message := &WebSocketMessage{
		Type:      MessageTypeServerUpdate,
		ID:        "test-id",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"cpu": 50.0},
		Metadata:  map[string]interface{}{"source": "monitor"},
	}

	// Serialize
	data, err := json.Marshal(message)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Deserialize
	var decodedMessage WebSocketMessage
	err = json.Unmarshal(data, &decodedMessage)
	require.NoError(t, err)

	assert.Equal(t, message.Type, decodedMessage.Type)
	assert.Equal(t, message.ID, decodedMessage.ID)
	assert.Equal(t, message.Data, decodedMessage.Data)
	assert.Equal(t, message.Metadata, decodedMessage.Metadata)
}

func TestAuthRequest_JSONSerialization(t *testing.T) {
	// Test auth request serialization
	authReq := &AuthRequest{
		AccessKey: "test-access-key",
	}

	// Serialize
	data, err := json.Marshal(authReq)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Deserialize
	var decodedAuthReq AuthRequest
	err = json.Unmarshal(data, &decodedAuthReq)
	require.NoError(t, err)

	assert.Equal(t, authReq.AccessKey, decodedAuthReq.AccessKey)
}

func TestSubscribeRequest_JSONSerialization(t *testing.T) {
	// Test subscribe request serialization
	subscribeReq := &SubscribeRequest{
		Events:   []MessageType{MessageTypeServerUpdate, MessageTypeAlert},
		Servers:  []string{"server1", "server2"},
		Projects: []string{"project1"},
	}

	// Serialize
	data, err := json.Marshal(subscribeReq)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Deserialize
	var decodedSubscribeReq SubscribeRequest
	err = json.Unmarshal(data, &decodedSubscribeReq)
	require.NoError(t, err)

	assert.Equal(t, subscribeReq.Events, decodedSubscribeReq.Events)
	assert.Equal(t, subscribeReq.Servers, decodedSubscribeReq.Servers)
	assert.Equal(t, subscribeReq.Projects, decodedSubscribeReq.Projects)
}

func TestAuthServiceAdapter(t *testing.T) {
	// Setup mock access key repository
	mockAccessKeyRepo := &MockAccessKeyRepository{}

	// Create adapter
	adapter := &authServiceAdapter{
		authService: mockAccessKeyRepo,
	}

	// Test successful validation
	ctx := context.Background()
	mockAccessKeyRepo.On("ValidateAccessKey", ctx, "valid-key").Return("test-project", nil)

	req := &AuthRequest{AccessKey: "valid-key"}
	result, err := adapter.ValidateAccessKey(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "test-project", result.ProjectKey)
	assert.Contains(t, result.Permissions, "read")
	assert.Contains(t, result.Permissions, "websocket")

	// Test failed validation
	mockAccessKeyRepo.On("ValidateAccessKey", ctx, "invalid-key").Return("", assert.AnError)

	req = &AuthRequest{AccessKey: "invalid-key"}
	result, err = adapter.ValidateAccessKey(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "invalid access key", result.Message)
	assert.Empty(t, result.ProjectKey)
	assert.Empty(t, result.Permissions)

	mockAccessKeyRepo.AssertExpectations(t)
}

// Integration test with actual WebSocket connections
func TestWebSocketService_HandleConnection_Integration(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockAuthService := &MockAuthService{}
	mockLogger := &MockLogger{}

	// Create service with no auth required for easier testing
	config := &Config{
		RequireAuth:   false,
		AuthTimeout:   1 * time.Second,
		ReadTimeout:   2 * time.Second,
		WriteTimeout:  1 * time.Second,
		PingPeriod:    500 * time.Millisecond,
		PongWait:      1 * time.Second,
		StatsInterval: 100 * time.Millisecond,
	}

	service := NewWebSocketService(
		mockServerRepo,
		mockCacheRepo,
		mockAccessKeyRepo,
		mockAuthService,
		mockLogger,
		config,
	)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
		conn, err := service.upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Failed to upgrade connection: %v", err)
			return
		}

		// Handle the connection
		_, err = service.HandleConnection(context.Background(), conn, r.RemoteAddr, r.UserAgent())
		if err != nil {
			t.Errorf("Failed to handle connection: %v", err)
		}
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Test message reception (should receive auth request if auth required, or system message)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	var wsMessage WebSocketMessage
	err = json.Unmarshal(message, &wsMessage)
	require.NoError(t, err)

	// Verify we got a message from the server
	assert.NotEmpty(t, wsMessage.Type)

	// Shutdown service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Shutdown(ctx)
	assert.NoError(t, err)
}