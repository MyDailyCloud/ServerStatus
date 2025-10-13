package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Mock repositories for testing
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

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
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
func TestAuthService_ValidateAccessKey_Success(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	config.CacheEnabled = false // Disable cache for this test
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("ValidateAccessKey", mock.Anything, "test-access-key").Return("test-project", nil)

	// Create auth request
	req := &AuthRequest{
		AccessKey:  "test-access-key",
		ProjectKey: "test-project",
		Permission: PermissionRead,
	}

	// Execute validation
	result, err := service.ValidateAccessKey(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "test-project", result.ProjectKey)
	assert.Contains(t, result.Permissions, PermissionRead)
	assert.Equal(t, config.TokenExpiration, result.TTL)

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestAuthService_ValidateAccessKey_AdminKey(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service with admin key
	config := DefaultConfig()
	config.AdminAccessKey = "admin-secret-key"
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Create auth request with admin key
	req := &AuthRequest{
		AccessKey:  "admin-secret-key",
		ProjectKey: "any-project",
		Permission: PermissionAdmin,
	}

	// Execute validation
	result, err := service.ValidateAccessKey(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "*", result.ProjectKey)
	assert.Contains(t, result.Permissions, PermissionAdmin)
	assert.Equal(t, "admin access granted", result.Message)

	// Verify no repository calls were made
	mockAccessKeyRepo.AssertNotCalled(t, "ValidateAccessKey")
}

func TestAuthService_ValidateAccessKey_InvalidKey(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	config.CacheEnabled = false
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("ValidateAccessKey", mock.Anything, "invalid-key").Return("", assert.AnError)

	// Create auth request
	req := &AuthRequest{
		AccessKey: "invalid-key",
	}

	// Execute validation
	result, err := service.ValidateAccessKey(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "invalid access key", result.Message)

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestAuthService_ValidateAccessKey_ProjectMismatch(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	config.CacheEnabled = false
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("ValidateAccessKey", mock.Anything, "test-access-key").Return("different-project", nil)

	// Create auth request
	req := &AuthRequest{
		AccessKey:  "test-access-key",
		ProjectKey: "test-project", // Different from repository result
	}

	// Execute validation
	result, err := service.ValidateAccessKey(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "does not have permission for this project")

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
}

func TestAuthService_ValidateAccessKey_CachedResult(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup cached result
	cachedResult := &AuthResult{
		Success:     true,
		ProjectKey:  "test-project",
		Permissions: []Permission{PermissionRead},
		TTL:         config.TokenExpiration,
	}

	// Setup mock expectations
	mockCacheRepo.On("Get", mock.Anything, "auth:result:test-access-key", mock.AnythingOfType("*auth.AuthResult")).Return(nil).Run(func(args mock.Arguments) {
		result := args.Get(2).(*AuthResult)
		*result = *cachedResult
	})
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Create auth request
	req := &AuthRequest{
		AccessKey: "test-access-key",
	}

	// Execute validation
	result, err := service.ValidateAccessKey(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "test-project", result.ProjectKey)

	// Verify cache was used, not repository
	mockAccessKeyRepo.AssertNotCalled(t, "ValidateAccessKey")
	mockCacheRepo.AssertExpectations(t)
}

func TestAuthService_GenerateAccessKey(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("GenerateAccessKey", mock.Anything, "server-key", "test-project").Return("generated-access-key", nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Execute generation
	accessKey, err := service.GenerateAccessKey(context.Background(), "server-key", "test-project")

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "generated-access-key", accessKey)

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthService_SaveAccessKey(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("SaveAccessKey", mock.Anything, "cache-key", "access-key").Return(nil)
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything)

	// Execute save
	err := service.SaveAccessKey(context.Background(), "cache-key", "access-key")

	// Assertions
	require.NoError(t, err)

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthService_DeleteAccessKey(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("DeleteAccessKey", mock.Anything, "access-key").Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "auth:result:access-key").Return(nil)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Execute delete
	err := service.DeleteAccessKey(context.Background(), "access-key")

	// Assertions
	require.NoError(t, err)

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthService_CleanupExpiredKeys(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockAccessKeyRepo.On("CleanupExpiredKeys", mock.Anything, config.TokenExpiration).Return(nil)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Execute cleanup
	err := service.CleanupExpiredKeys(context.Background())

	// Assertions
	require.NoError(t, err)

	// Verify mock expectations
	mockAccessKeyRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthService_ValidatePermissionString(t *testing.T) {
	service := &AuthService{}

	tests := []struct {
		permission string
		expected   bool
	}{
		{"read", true},
		{"write", true},
		{"delete", true},
		{"export", true},
		{"admin", true},
		{"websocket", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			result := service.ValidatePermissionString(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthService_GetDefaultPermissions(t *testing.T) {
	service := &AuthService{}
	permissions := service.GetDefaultPermissions()

	assert.Contains(t, permissions, PermissionRead)
	assert.Contains(t, permissions, PermissionWebSocket)
	assert.Len(t, permissions, 2)
}

func TestAuthService_GetAllPermissions(t *testing.T) {
	service := &AuthService{}
	permissions := service.GetAllPermissions()

	expectedPermissions := []Permission{
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionExport,
		PermissionAdmin,
		PermissionWebSocket,
	}

	assert.Equal(t, expectedPermissions, permissions)
}

func TestAuthService_HasPermission(t *testing.T) {
	service := &AuthService{}

	tests := []struct {
		name        string
		permissions []Permission
		required    Permission
		expected    bool
	}{
		{
			name:        "has direct permission",
			permissions: []Permission{PermissionRead, PermissionWrite},
			required:    PermissionRead,
			expected:    true,
		},
		{
			name:        "has admin permission",
			permissions: []Permission{PermissionAdmin},
			required:    PermissionRead,
			expected:    true,
		},
		{
			name:        "no permission",
			permissions: []Permission{PermissionRead},
			required:    PermissionWrite,
			expected:    false,
		},
		{
			name:        "empty permissions",
			permissions: []Permission{},
			required:    PermissionRead,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.HasPermission(tt.permissions, tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthService_GetAuthStats(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	config := DefaultConfig()
	config.CacheEnabled = true
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	// Setup mock expectations
	mockCacheRepo.On("GetStats", mock.Anything).Return(map[string]interface{}{
		"hits":   100,
		"misses": 20,
	}, nil)

	// Execute stats
	stats, err := service.GetAuthStats(context.Background())

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, true, stats["cache_enabled"])
	assert.Equal(t, config.TokenExpiration.String(), stats["token_expiration"])
	assert.Equal(t, config.MaxFailedAttempts, stats["max_failed_attempts"])
	assert.Equal(t, config.LockoutDuration.String(), stats["lockout_duration"])

	cacheStats, ok := stats["cache_stats"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 100, cacheStats["hits"])
	assert.Equal(t, 20, cacheStats["misses"])

	// Verify mock expectations
	mockCacheRepo.AssertExpectations(t)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 24*time.Hour, config.TokenExpiration)
	assert.True(t, config.CacheEnabled)
	assert.Equal(t, 15*time.Minute, config.CacheTTL)
	assert.Equal(t, 5, config.MaxFailedAttempts)
	assert.Equal(t, 15*time.Minute, config.LockoutDuration)
	assert.Empty(t, config.AdminAccessKey)
}

func TestNewAuthService(t *testing.T) {
	// Setup mocks
	mockAccessKeyRepo := &MockAccessKeyRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Test with custom config
	config := &Config{
		TokenExpiration: 12 * time.Hour,
		CacheEnabled:    false,
	}
	service := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, config)

	assert.Equal(t, mockAccessKeyRepo, service.accessKeyRepo)
	assert.Equal(t, mockCacheRepo, service.cacheRepo)
	assert.Equal(t, mockLogger, service.logger)
	assert.Equal(t, config, service.config)

	// Test with nil config (should use default)
	service2 := NewAuthService(mockAccessKeyRepo, mockCacheRepo, mockLogger, nil)
	defaultConfig := DefaultConfig()

	assert.Equal(t, defaultConfig.TokenExpiration, service2.config.TokenExpiration)
	assert.Equal(t, defaultConfig.CacheEnabled, service2.config.CacheEnabled)
}
