package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
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

func (m *MockServerRepository) GetServersByProject(ctx context.Context, projectKey string, page *repository.Pagination) ([]*models.ServerInfo, error) {
	args := m.Called(ctx, projectKey, page)
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

func (m *MockAccessKeyRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAccessKeyRepository) Close() error {
	args := m.Called()
	return args.Error(0)
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

// Test cases
func TestHealthService_CheckHealth_AllHealthy(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Setup mock expectations
	mockServerRepo.On("Ping").Return(nil)
	mockCacheRepo.On("IsConnected").Return(true)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*string)
		*dest = "ok"
	})
	mockCacheRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("GetType").Return("redis")
	mockAccessKeyRepo.On("GenerateAccessKey", mock.Anything, mock.Anything, mock.Anything).Return("test-key", nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Execute health check
	req := &CheckRequest{
		IncludeSystem: true,
		IncludeMemory: true,
		CheckAll:      true,
	}
	result, err := service.CheckHealth(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Len(t, result.Components, 5) // database, cache, auth, system, memory
	assert.Equal(t, "2.1.0", result.Version)
	assert.NotNil(t, result.Memory)

	// Check component statuses
	assert.Equal(t, StatusHealthy, result.Components["database"].Status)
	assert.Equal(t, StatusHealthy, result.Components["cache"].Status)
	assert.Equal(t, StatusHealthy, result.Components["auth"].Status)
	assert.Equal(t, StatusHealthy, result.Components["system"].Status)
	assert.Equal(t, StatusHealthy, result.Components["memory"].Status)

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestHealthService_CheckHealth_DatabaseUnhealthy(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Setup mock expectations - database fails
	mockServerRepo.On("Ping").Return(assert.AnError)
	mockCacheRepo.On("IsConnected").Return(true)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*string)
		*dest = "ok"
	})
	mockCacheRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("GetType").Return("redis")
	mockAccessKeyRepo.On("GenerateAccessKey", mock.Anything, mock.Anything, mock.Anything).Return("test-key", nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Execute health check
	req := &CheckRequest{CheckAll: true}
	result, err := service.CheckHealth(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusDegraded, result.Status) // Some components are healthy
	assert.Equal(t, StatusUnhealthy, result.Components["database"].Status)
	assert.Contains(t, result.Components["database"].Message, "database connection failed")

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestHealthService_CheckHealth_CacheDisconnected(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Setup mock expectations - cache disconnected
	mockServerRepo.On("Ping").Return(nil)
	mockCacheRepo.On("IsConnected").Return(false)
	mockAccessKeyRepo.On("GenerateAccessKey", mock.Anything, mock.Anything, mock.Anything).Return("test-key", nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Execute health check
	req := &CheckRequest{CheckAll: true}
	result, err := service.CheckHealth(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusDegraded, result.Status)
	assert.Equal(t, StatusDegraded, result.Components["cache"].Status)
	assert.Contains(t, result.Components["cache"].Message, "cache service is not connected")

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestHealthService_IsHealthy_True(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Setup mock expectations - IsHealthy only checks database, cache, and auth
	mockServerRepo.On("Ping").Return(nil)
	mockCacheRepo.On("IsConnected").Return(true)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*string)
		*dest = "ok"
	})
	mockCacheRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("GetType").Return("redis")
	mockAccessKeyRepo.On("GenerateAccessKey", mock.Anything, mock.Anything, mock.Anything).Return("test-key", nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Execute quick health check
	result := service.IsHealthy(context.Background())

	// Assertions
	assert.True(t, result)

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestHealthService_IsHealthy_False(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Setup mock expectations - database fails
	mockServerRepo.On("Ping").Return(assert.AnError)
	mockCacheRepo.On("IsConnected").Return(true)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*string)
		*dest = "ok"
	})
	mockCacheRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("GetType").Return("redis")
	mockAccessKeyRepo.On("GenerateAccessKey", mock.Anything, mock.Anything, mock.Anything).Return("test-key", nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Execute quick health check
	result := service.IsHealthy(context.Background())

	// Assertions
	assert.False(t, result)

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestHealthService_GetComponentStatus(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Test database component
	mockServerRepo.On("Ping").Return(nil)
	status, err := service.GetComponentStatus(context.Background(), "database")

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "database", status.Name)
	assert.Equal(t, StatusHealthy, status.Status)

	// Test unknown component
	status, err = service.GetComponentStatus(context.Background(), "unknown")
	require.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "unknown component")

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
}

func TestHealthService_GetServiceInfo(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	config.Version = "test-version"
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Execute
	info := service.GetServiceInfo(context.Background())

	// Assertions
	require.NotNil(t, info)
	assert.Equal(t, "ServerStatus Data Server", info["name"])
	assert.Equal(t, "test-version", info["version"])
	assert.Contains(t, info, "uptime")
	assert.Contains(t, info, "start_time")
	assert.Contains(t, info, "go_version")
}

func TestHealthService_UpdateConfig(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, config)

	// Test updating config
	newConfig := &Config{
		Version:         "new-version",
		MemoryThreshold: 1024,
	}
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	service.UpdateConfig(newConfig)

	// Verify config was updated
	updatedConfig := service.GetConfig()
	assert.Equal(t, "new-version", updatedConfig.Version)
	assert.Equal(t, uint64(1024), updatedConfig.MemoryThreshold)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.CheckInterval)
	assert.Equal(t, 5*time.Second, config.Timeout)
	assert.Equal(t, uint64(512), config.MemoryThreshold)
	assert.Equal(t, 100*time.Millisecond, config.ResponseThreshold)
	assert.Equal(t, "2.1.0", config.Version)
}

func TestNewHealthService(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Test with custom config
	customConfig := &Config{
		Version: "custom-version",
	}
	service := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, customConfig)

	assert.Equal(t, mockServerRepo, service.serverRepo)
	assert.Equal(t, mockCacheRepo, service.cacheRepo)
	assert.Equal(t, mockAccessKeyRepo, service.accessKeyRepo)
	assert.Equal(t, mockLogger, service.logger)
	assert.Equal(t, customConfig, service.config)

	// Test with nil config (should use default)
	service2 := NewHealthService(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger, nil)
	defaultConfig := DefaultConfig()

	assert.Equal(t, defaultConfig.Enabled, service2.config.Enabled)
	assert.Equal(t, defaultConfig.Version, service2.config.Version)
}

func TestFactory_CreateService(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockLogger := &MockLogger{}

	// Create factory
	factory := NewFactory(mockServerRepo, mockCacheRepo, mockAccessKeyRepo, mockLogger)

	// Test create service with custom config
	customConfig := &Config{
		Version: "factory-test",
	}
	service := factory.CreateService(customConfig)

	assert.NotNil(t, service)
	assert.Equal(t, customConfig, service.config)

	// Test create default service
	defaultService := factory.CreateDefaultService()
	assert.NotNil(t, defaultService)
	assert.Equal(t, DefaultConfig().Version, defaultService.config.Version)
}
