package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServerRepository 模拟服务器仓库
type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) GetServersByProject(ctx context.Context, projectKey string, pagination *repository.Pagination) ([]*models.ServerInfo, error) {
	args := m.Called(ctx, projectKey, pagination)
	return args.Get(0).([]*models.ServerInfo), args.Error(1)
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

// MockHistoryRepository 模拟历史数据仓库
type MockHistoryRepository struct {
	mock.Mock
}

func (m *MockHistoryRepository) SaveHistoryData(ctx context.Context, data *models.SystemInfo) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockHistoryRepository) GetHostHistory(ctx context.Context, hostname, projectKey string, limit int) ([]*models.HistoryData, error) {
	args := m.Called(ctx, hostname, projectKey, limit)
	return args.Get(0).([]*models.HistoryData), args.Error(1)
}

func (m *MockHistoryRepository) GetHistoryByTimeRange(ctx context.Context, hostname, projectKey string, start, end time.Time) ([]*models.HistoryData, error) {
	args := m.Called(ctx, hostname, projectKey, start, end)
	return args.Get(0).([]*models.HistoryData), args.Error(1)
}

func (m *MockHistoryRepository) CleanupOldData(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockHistoryRepository) GetHistoryCount(ctx context.Context, hostname, projectKey string) (int, error) {
	args := m.Called(ctx, hostname, projectKey)
	return args.Int(0), args.Error(1)
}

func (m *MockHistoryRepository) GetAggregatedData(ctx context.Context, hostname, projectKey string, interval time.Duration, limit int) ([]*models.HistoryData, error) {
	args := m.Called(ctx, hostname, projectKey, interval, limit)
	return args.Get(0).([]*models.HistoryData), args.Error(1)
}

func (m *MockHistoryRepository) GetHistoryByTimeRangePaged(ctx context.Context, hostname, projectKey string, start, end time.Time, pagination *repository.Pagination) ([]*models.HistoryData, int, error) {
	args := m.Called(ctx, hostname, projectKey, start, end, pagination)
	return args.Get(0).([]*models.HistoryData), args.Int(1), args.Error(2)
}

func (m *MockHistoryRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

// MockCacheRepository 模拟缓存仓库
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

func (m *MockCacheRepository) HealthChecker() repository.HealthChecker {
	args := m.Called()
	return args.Get(0).(repository.HealthChecker)
}

// TestServerService_RegisterServer 测试注册服务器
func TestServerService_RegisterServer(t *testing.T) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	// 准备测试数据
	systemInfo := &models.SystemInfo{
		SessionID:  "test-session-123",
		Hostname:   "test-server",
		ProjectKey: "test-project",
		Timestamp:  time.Now(),
		OS: models.OSInfo{
			Platform:     "Linux",
			Architecture: "x86_64",
			Hostname:     "test-server",
		},
		CPU: models.CPUInfo{
			CoreCount:    4,
			UsagePercent: 12.3,
		},
		Memory: models.MemInfo{
			Total: 8589934592, // 8GB
			Used:  2147483648,
			Free:  6442450944,
		},
		Disk: models.DiskInfo{
			Total: 107374182400, // 100GB
			Used:  21474836480,
			Free:  85900299520,
		},
	}

	// 设置模拟期望
	mockServerRepo.On("CreateServer", mock.Anything, mock.Anything).Return(nil)
	mockHistoryRepo.On("SaveHistoryData", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	err := service.RegisterServer(context.Background(), systemInfo)

	// 验证结果
	assert.NoError(t, err)
	mockServerRepo.AssertExpectations(t)
	mockHistoryRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
}

// TestServerService_RegisterServer_InvalidData 测试注册服务器时的无效数据
func TestServerService_RegisterServer_InvalidData(t *testing.T) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	// 测试用例
	testCases := []struct {
		name    string
		info    *models.SystemInfo
		wantErr bool
	}{
		{
			name: "empty session ID",
			info: &models.SystemInfo{
				Hostname:  "test-server",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty hostname",
			info: &models.SystemInfo{
				SessionID: "test-session-123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty timestamp",
			info: &models.SystemInfo{
				SessionID: "test-session-123",
				Hostname:  "test-server",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.RegisterServer(context.Background(), tc.info)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServerService_GetServer 测试获取服务器信息
func TestServerService_GetServer(t *testing.T) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	sessionID := "test-session-123"
	expectedServer := &models.ServerInfo{
		SessionID: sessionID,
		Hostname:  "test-server",
	}

	// 测试缓存未命中
	mockCacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("cache miss"))
	mockServerRepo.On("GetServer", mock.Anything, sessionID).Return(expectedServer, nil)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	result, err := service.GetServer(context.Background(), sessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, expectedServer, result)
	mockCacheRepo.AssertExpectations(t)
	mockServerRepo.AssertExpectations(t)
}

// TestServerService_GetServers 测试获取服务器列表
func TestServerService_GetServers(t *testing.T) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	// 准备测试数据
	expectedServers := []*models.ServerInfo{
		{SessionID: "test-1", Hostname: "server-1"},
		{SessionID: "test-2", Hostname: "server-2"},
	}

	filter := &repository.ServerFilter{
		ProjectKey: "test-project",
		Page:       1,
		PageSize:   10,
	}

	// 设置模拟期望
	mockServerRepo.On("GetServersByProject", mock.Anything, "test-project", mock.Anything).Return(expectedServers, nil)

	// 执行测试
	result, err := service.GetServers(context.Background(), filter)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockServerRepo.AssertExpectations(t)
}

// TestServerService_GetProjectStats 测试获取项目统计
func TestServerService_GetProjectStats(t *testing.T) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	projectKey := "test-project"

	// 准备测试数据
	onlineServers := []*models.ServerInfo{
		{
			SessionID: "test-1",
			Hostname:  "server-1",
			Latest:    &models.SystemInfo{CPU: models.CPUInfo{UsagePercent: 50.0}, Memory: models.MemInfo{UsagePercent: 60.0}, Disk: models.DiskInfo{UsagePercent: 70.0}},
		},
		{
			SessionID: "test-2",
			Hostname:  "server-2",
			Latest:    &models.SystemInfo{CPU: models.CPUInfo{UsagePercent: 30.0}, Memory: models.MemInfo{UsagePercent: 40.0}, Disk: models.DiskInfo{UsagePercent: 50.0}},
		},
	}

	// 设置模拟期望
	mockServerRepo.On("GetServerCount", mock.Anything, projectKey).Return(3, nil)
	mockServerRepo.On("GetOnlineServers", mock.Anything, projectKey, mock.Anything).Return(onlineServers, nil)

	// 执行测试
	stats, err := service.GetProjectStats(context.Background(), projectKey)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.TotalServers)
	assert.Equal(t, 2, stats.OnlineServers)
	assert.Equal(t, 1, stats.OfflineServers)
	assert.Equal(t, 40.0, stats.AvgCPUUsage)    // (50+30)/2
	assert.Equal(t, 50.0, stats.AvgMemoryUsage) // (60+40)/2
	assert.Equal(t, 60.0, stats.AvgDiskUsage)   // (70+50)/2

	mockServerRepo.AssertExpectations(t)
}

// TestServerService_RemoveServer 测试移除服务器
func TestServerService_RemoveServer(t *testing.T) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	sessionID := "test-session-123"

	// 设置模拟期望
	mockServerRepo.On("DeleteServer", mock.Anything, sessionID).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	err := service.RemoveServer(context.Background(), sessionID)

	// 验证结果
	assert.NoError(t, err)
	mockServerRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
}

// TestCalculateAverageUsage 测试计算平均资源使用率
func TestCalculateAverageUsage(t *testing.T) {
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	service := NewServerService(nil, nil, nil, testLogger)

	// 测试用例
	testCases := []struct {
		name     string
		servers  []*models.ServerInfo
		wantCPU  float64
		wantMem  float64
		wantDisk float64
	}{
		{
			name:     "no servers",
			servers:  []*models.ServerInfo{},
			wantCPU:  0,
			wantMem:  0,
			wantDisk: 0,
		},
		{
			name: "single server",
			servers: []*models.ServerInfo{
				{
					Latest: &models.SystemInfo{
						CPU:    models.CPUInfo{UsagePercent: 50.0},
						Memory: models.MemInfo{UsagePercent: 60.0},
						Disk:   models.DiskInfo{UsagePercent: 70.0},
					},
				},
			},
			wantCPU:  50.0,
			wantMem:  60.0,
			wantDisk: 70.0,
		},
		{
			name: "multiple servers",
			servers: []*models.ServerInfo{
				{
					Latest: &models.SystemInfo{
						CPU:    models.CPUInfo{UsagePercent: 50.0},
						Memory: models.MemInfo{UsagePercent: 60.0},
						Disk:   models.DiskInfo{UsagePercent: 70.0},
					},
				},
				{
					Latest: &models.SystemInfo{
						CPU:    models.CPUInfo{UsagePercent: 30.0},
						Memory: models.MemInfo{UsagePercent: 40.0},
						Disk:   models.DiskInfo{UsagePercent: 50.0},
					},
				},
			},
			wantCPU:  40.0,
			wantMem:  50.0,
			wantDisk: 60.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpu, mem, disk := service.calculateAverageUsage(tc.servers)
			assert.Equal(t, tc.wantCPU, cpu)
			assert.Equal(t, tc.wantMem, mem)
			assert.Equal(t, tc.wantDisk, disk)
		})
	}
}

// BenchmarkServerService_RegisterServer 基准测试注册服务器
func BenchmarkServerService_RegisterServer(b *testing.B) {
	// 创建模拟对象
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}

	// 创建测试日志器
	loggerConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(loggerConfig)

	// 创建服务
	service := NewServerService(mockServerRepo, mockHistoryRepo, mockCacheRepo, testLogger)

	// 准备测试数据
	systemInfo := &models.SystemInfo{
		SessionID:  "test-session-123",
		Hostname:   "test-server",
		ProjectKey: "test-project",
		Timestamp:  time.Now(),
		OS: models.OSInfo{
			Platform:     "Linux",
			Architecture: "x86_64",
			Hostname:     "test-server",
		},
		CPU: models.CPUInfo{
			CoreCount:    4,
			UsagePercent: 10,
		},
		Memory: models.MemInfo{
			Total: 8589934592,
			Used:  2147483648,
			Free:  6442450944,
		},
		Disk: models.DiskInfo{
			Total: 107374182400,
			Used:  21474836480,
			Free:  85900299520,
		},
	}

	// 设置模拟期望
	mockServerRepo.On("CreateServer", mock.Anything, mock.Anything).Return(nil)
	mockHistoryRepo.On("SaveHistoryData", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 修改sessionID避免重复
		systemInfo.SessionID = fmt.Sprintf("test-session-%d", i)
		_ = service.RegisterServer(context.Background(), systemInfo)
	}
}
