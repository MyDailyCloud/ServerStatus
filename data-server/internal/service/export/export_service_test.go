package export

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// testify模拟仓库与日志
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

func (m *MockServerRepository) GetServersByProject(ctx context.Context, projectKey string, pagination *repository.Pagination) ([]*models.ServerInfo, error) {
	args := m.Called(ctx, projectKey, pagination)
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

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) { m.Called(args...) }
func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}
func (m *MockLogger) Info(args ...interface{}) { m.Called(args...) }
func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}
func (m *MockLogger) Warn(args ...interface{}) { m.Called(args...) }
func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}
func (m *MockLogger) Error(args ...interface{}) { m.Called(args...) }
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}
func (m *MockLogger) Fatal(args ...interface{}) { m.Called(args...) }
func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}
func (m *MockLogger) Panic(args ...interface{}) { m.Called(args...) }
func (m *MockLogger) Panicf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}
func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	m.Called(key, value)
	return m
}
func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	m.Called(fields)
	return m
}
func (m *MockLogger) WithError(err error) logger.Logger {
	m.Called(err)
	return m
}

type mockHistoryRepo struct{ data []*models.HistoryData }

func (m *mockHistoryRepo) GetHistoryByTimeRange(ctx context.Context, hostname, projectKey string, start, end time.Time) ([]*models.HistoryData, error) {
	return m.data, nil
}
func (m *mockHistoryRepo) SaveHistoryData(ctx context.Context, data *models.SystemInfo) error {
	return nil
}
func (m *mockHistoryRepo) GetHostHistory(ctx context.Context, hostname, projectKey string, limit int) ([]*models.HistoryData, error) {
	return nil, nil
}
func (m *mockHistoryRepo) CleanupOldData(ctx context.Context, before time.Time) error { return nil }
func (m *mockHistoryRepo) GetHistoryCount(ctx context.Context, hostname, projectKey string) (int, error) {
	return len(m.data), nil
}
func (m *mockHistoryRepo) GetAggregatedData(ctx context.Context, hostname, projectKey string, interval time.Duration, limit int) ([]*models.HistoryData, error) {
	return nil, nil
}
func (m *mockHistoryRepo) GetHistoryByTimeRangePaged(ctx context.Context, hostname, projectKey string, start, end time.Time, pagination *repository.Pagination) ([]*models.HistoryData, int, error) {
	return m.data, len(m.data), nil
}
func (m *mockHistoryRepo) Ping() error { return nil }

type mockCacheRepo struct{}

func (m *mockCacheRepo) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockCacheRepo) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *mockCacheRepo) Delete(ctx context.Context, key string) error                { return nil }
func (m *mockCacheRepo) Exists(ctx context.Context, key string) (bool, error)        { return false, nil }
func (m *mockCacheRepo) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockCacheRepo) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *mockCacheRepo) DeleteMultiple(ctx context.Context, keys []string) error { return nil }
func (m *mockCacheRepo) ClearPattern(ctx context.Context, pattern string) error  { return nil }
func (m *mockCacheRepo) Keys(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}
func (m *mockCacheRepo) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *mockCacheRepo) IsConnected() bool { return true }
func (m *mockCacheRepo) GetType() string   { return "memory" }

func newTestLogger(t *testing.T) logger.Logger {
	log, err := logger.NewLogger(&logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	})
	require.NoError(t, err)
	return log
}

func sampleHistoryData() []*models.HistoryData {
	ts := time.Now().Add(-1 * time.Minute)
	return []*models.HistoryData{
		{
			Timestamp:       ts,
			Hostname:        "host-a",
			SessionID:       "session-a",
			ProjectKey:      "proj-1",
			CPUUsage:        25.5,
			MemoryUsed:      2 * 1024 * 1024 * 1024,
			MemoryUsage:     50,
			MemoryAvailable: 2 * 1024 * 1024 * 1024,
			DiskUsed:        10 * 1024 * 1024 * 1024,
			DiskUsage:       20,
			DiskAvailable:   40 * 1024 * 1024 * 1024,
			NetworkRx:       1234,
			NetworkTx:       5678,
		},
	}
}

func TestConvertHistoryDataToSystemInfo(t *testing.T) {
	log := newTestLogger(t)
	svc := NewExportService(nil, &mockHistoryRepo{}, &mockCacheRepo{}, log)

	h := sampleHistoryData()[0]
	sys := svc.convertHistoryDataToSystemInfo(h)

	assert.Equal(t, h.Hostname, sys.Hostname)
	assert.Equal(t, h.CPUUsage, sys.CPU.UsagePercent)
	assert.Equal(t, h.MemoryUsed, sys.Memory.Used)
	assert.Equal(t, h.DiskUsed, sys.Disk.Used)
	assert.Equal(t, h.NetworkRx, sys.Network.BytesRecv)
	assert.Equal(t, h.NetworkTx, sys.Network.BytesSent)
}

func TestExportServersJSONWithHostnames(t *testing.T) {
	log := newTestLogger(t)
	svc := NewExportService(nil, &mockHistoryRepo{data: sampleHistoryData()}, &mockCacheRepo{}, log)

	req := &ExportRequest{
		ProjectKey:   "proj-1",
		Hostnames:    []string{"host-a"},
		StartTime:    time.Now().Add(-2 * time.Hour),
		EndTime:      time.Now(),
		Format:       FormatJSON,
		IncludeTypes: []ExportDataType{DataTypeSystemInfo},
		Limit:        10,
		Offset:       0,
	}

	result, err := svc.ExportServers(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "application/json", result.ContentType)
	assert.Equal(t, 1, result.RecordCount)

	body, err := io.ReadAll(result.Data)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"hostname":"host-a"`)
}

func TestExportHistoryJSONWithHostnames(t *testing.T) {
	log := newTestLogger(t)
	svc := NewExportService(nil, &mockHistoryRepo{data: sampleHistoryData()}, &mockCacheRepo{}, log)

	req := &ExportRequest{
		ProjectKey:   "proj-1",
		Hostnames:    []string{"host-a"},
		StartTime:    time.Now().Add(-2 * time.Hour),
		EndTime:      time.Now(),
		Format:       FormatJSON,
		IncludeTypes: []ExportDataType{DataTypeHistory},
		Limit:        10,
		Offset:       0,
	}

	result, err := svc.ExportHistory(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "application/json", result.ContentType)
	assert.Equal(t, 1, result.RecordCount)

	body, err := io.ReadAll(result.Data)
	require.NoError(t, err)

	dec := json.NewDecoder(bytes.NewReader(body))
	var arr []map[string]interface{}
	require.NoError(t, dec.Decode(&arr))
	assert.Equal(t, "host-a", arr[0]["hostname"])
}
func createTestServerInfo() *models.ServerInfo {
	now := time.Now()
	return &models.ServerInfo{
		SessionID:   "test-session-1",
		Hostname:    "test-host",
		ProjectKey:  "test-project",
		OS:          "linux",
		Arch:        "amd64",
		CPUCores:    4,
		MemoryTotal: 8589934592,   // 8GB
		DiskTotal:   107374182400, // 100GB
		Uptime:      3600,
		BootTime:    now.Add(-time.Hour),
		CreatedAt:   now.Add(-2 * time.Hour),
		UpdatedAt:   now,
		Latest: &models.SystemInfo{
			Hostname:  "test-host",
			SessionID: "test-session-1",
			Timestamp: now,
			CPU: models.CPUInfo{
				UsagePercent: 25.5,
				CoreCount:    4,
			},
			Memory: models.MemInfo{
				Used:  4294967296,
				Free:  4294967296,
				Total: 8589934592,
			},
			Disk: models.DiskInfo{
				Used:  53687091200,
				Free:  53687091200,
				Total: 107374182400,
			},
			Network: models.NetInfo{
				BytesRecv: 1073741824,
				BytesSent: 2147483648,
			},
		},
	}
}

func createTestHistoryData() []*models.HistoryData {
	now := time.Now()
	var history []*models.HistoryData

	for i := 0; i < 10; i++ {
		timestamp := now.Add(-time.Duration(i) * time.Minute)
		history = append(history, &models.HistoryData{
			Hostname:        "test-host",
			SessionID:       "test-session-1",
			ProjectKey:      "test-project",
			Timestamp:       timestamp,
			CPUUsage:        float64(20 + i),
			MemoryUsed:      4294967296 + uint64(i*100000000),
			MemoryUsage:     float64(40 + i),
			MemoryAvailable: 4294967296 - uint64(i*100000000),
			DiskUsed:        53687091200 + uint64(i*1000000000),
			DiskUsage:       float64(10 + i),
			DiskAvailable:   53687091200 - uint64(i*1000000000),
			NetworkRx:       1073741824 + uint64(i*10000000),
			NetworkTx:       2147483648 + uint64(i*20000000),
		})
	}

	return history
}

// Test cases
func TestExportService_ExportServersCSV(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	service := NewExportService(mockServerRepo, mockHistoryRepo, mockCacheRepo, mockLogger)

	// Setup test data
	server := createTestServerInfo()
	servers := []*models.ServerInfo{server}

	// Setup mock expectations
	mockServerRepo.On("GetAllServers", mock.Anything, "test-project", 0, 10000).Return(servers, nil)

	// Create export request
	req := &ExportRequest{
		ProjectKey:   "test-project",
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		Format:       FormatCSV,
		IncludeTypes: []ExportDataType{DataTypeServerInfo, DataTypeSystemInfo},
		Limit:        10000,
		Offset:       0,
	}

	// Mock logger calls
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Execute export
	result, err := service.ExportServers(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "text/csv", result.ContentType)
	assert.Equal(t, 1, result.RecordCount)
	assert.True(t, strings.HasPrefix(result.Filename, "servers_export_test-project_"))
	assert.True(t, strings.HasSuffix(result.Filename, ".csv"))

	// Verify CSV content
	content, err := io.ReadAll(result.Data)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	assert.True(t, len(lines) >= 2) // Header + at least one data line

	// Verify header
	header := lines[0]
	expectedHeaders := []string{
		"SessionID", "Hostname", "ProjectKey", "OS", "Architecture",
		"CPUCores", "MemoryTotal", "DiskTotal", "Uptime", "BootTime",
		"CreatedAt", "UpdatedAt", "CPUUsage", "MemoryUsed", "MemoryAvailable",
		"DiskUsed", "DiskAvailable", "NetworkRx", "NetworkTx", "LoadAvg",
		"ProcessCount", "LastUpdate",
	}
	for _, expectedHeader := range expectedHeaders {
		assert.Contains(t, header, expectedHeader)
	}

	// Close data reader
	result.Data.Close()

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestExportService_ExportServersJSON(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	service := NewExportService(mockServerRepo, mockHistoryRepo, mockCacheRepo, mockLogger)

	// Setup test data
	server := createTestServerInfo()
	servers := []*models.ServerInfo{server}

	// Setup mock expectations
	mockServerRepo.On("GetAllServers", mock.Anything, "test-project", 0, 10000).Return(servers, nil)

	// Create export request
	req := &ExportRequest{
		ProjectKey:   "test-project",
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		Format:       FormatJSON,
		IncludeTypes: []ExportDataType{DataTypeServerInfo},
		Limit:        10000,
		Offset:       0,
	}

	// Mock logger calls
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Execute export
	result, err := service.ExportServers(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "application/json", result.ContentType)
	assert.Equal(t, 1, result.RecordCount)
	assert.True(t, strings.HasPrefix(result.Filename, "servers_export_test-project_"))
	assert.True(t, strings.HasSuffix(result.Filename, ".json"))

	// Verify JSON content
	content, err := io.ReadAll(result.Data)
	require.NoError(t, err)

	var exportData map[string]interface{}
	err = json.Unmarshal(content, &exportData)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, exportData, "metadata")
	assert.Contains(t, exportData, "servers")

	metadata := exportData["metadata"].(map[string]interface{})
	assert.Equal(t, "test-project", metadata["project_key"])
	assert.Equal(t, float64(1), metadata["total_servers"])

	serversData := exportData["servers"].([]interface{})
	assert.Len(t, serversData, 1)

	// Close data reader
	result.Data.Close()

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestExportService_ExportHistoryCSV(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	service := NewExportService(mockServerRepo, mockHistoryRepo, mockCacheRepo, mockLogger)

	// Setup test data
	history := createTestHistoryData()
	servers := []*models.ServerInfo{createTestServerInfo()}

	// Setup mock expectations
	mockServerRepo.On("GetAllServers", mock.Anything, "test-project", 0, 1000).Return(servers, nil)
	mockHistoryRepo.On("GetHistoryByTimeRange", mock.Anything, "test-session-1", "test-project", mock.Anything, mock.Anything).Return(history, nil)

	// Create export request
	req := &ExportRequest{
		ProjectKey: "test-project",
		StartTime:  time.Now().Add(-2 * time.Hour),
		EndTime:    time.Now(),
		Format:     FormatCSV,
		Limit:      10000,
		Offset:     0,
	}

	// Mock logger calls
	mockLogger.On("WithFields", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything)

	// Execute export
	result, err := service.ExportHistory(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "text/csv", result.ContentType)
	assert.Equal(t, len(history), result.RecordCount)
	assert.True(t, strings.HasPrefix(result.Filename, "history_export_test-project_"))
	assert.True(t, strings.HasSuffix(result.Filename, ".csv"))

	// Close data reader
	result.Data.Close()

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
	mockHistoryRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestExportService_ValidateRequest(t *testing.T) {
	service := &ExportService{}

	tests := []struct {
		name    string
		req     *ExportRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &ExportRequest{
				ProjectKey:   "test-project",
				StartTime:    time.Now().Add(-time.Hour),
				EndTime:      time.Now(),
				Format:       FormatCSV,
				IncludeTypes: []ExportDataType{DataTypeServerInfo},
				Limit:        1000,
			},
			wantErr: false,
		},
		{
			name: "missing project key",
			req: &ExportRequest{
				StartTime:    time.Now().Add(-time.Hour),
				EndTime:      time.Now(),
				Format:       FormatCSV,
				IncludeTypes: []ExportDataType{DataTypeServerInfo},
			},
			wantErr: true,
		},
		{
			name: "start time after end time",
			req: &ExportRequest{
				ProjectKey:   "test-project",
				StartTime:    time.Now(),
				EndTime:      time.Now().Add(-time.Hour),
				Format:       FormatCSV,
				IncludeTypes: []ExportDataType{DataTypeServerInfo},
			},
			wantErr: true,
		},
		{
			name: "time range too large",
			req: &ExportRequest{
				ProjectKey:   "test-project",
				StartTime:    time.Now().Add(-400 * 24 * time.Hour), // 400 days
				EndTime:      time.Now(),
				Format:       FormatCSV,
				IncludeTypes: []ExportDataType{DataTypeServerInfo},
			},
			wantErr: true,
		},
		{
			name: "limit too high",
			req: &ExportRequest{
				ProjectKey:   "test-project",
				StartTime:    time.Now().Add(-time.Hour),
				EndTime:      time.Now(),
				Format:       FormatCSV,
				IncludeTypes: []ExportDataType{DataTypeServerInfo},
				Limit:        200000, // Exceeds 100000 limit
			},
			wantErr: false, // Should be auto-corrected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportService_GetExportFormats(t *testing.T) {
	service := &ExportService{}
	formats := service.GetExportFormats()

	expectedFormats := []ExportFormat{FormatCSV, FormatJSON}
	assert.Equal(t, expectedFormats, formats)
}

func TestExportService_GetExportDataTypes(t *testing.T) {
	service := &ExportService{}
	types := service.GetExportDataTypes()

	expectedTypes := []ExportDataType{DataTypeServerInfo, DataTypeSystemInfo, DataTypeHistory}
	assert.Equal(t, expectedTypes, types)
}

func TestExportService_EstimateExportSize(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockLogger := &MockLogger{}

	// Create service
	service := NewExportService(mockServerRepo, mockHistoryRepo, mockCacheRepo, mockLogger)

	// Setup test data
	server := createTestServerInfo()
	servers := []*models.ServerInfo{server}

	// Setup mock expectations
	mockServerRepo.On("GetAllServers", mock.Anything, "test-project", 0, 10).Return(servers, nil)

	// Create export request
	req := &ExportRequest{
		ProjectKey:   "test-project",
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		Format:       FormatCSV,
		IncludeTypes: []ExportDataType{DataTypeServerInfo},
		Limit:        1000, // Large limit for estimation
		Offset:       0,
	}

	// Execute estimation
	size, err := service.EstimateExportSize(context.Background(), req)

	// Assertions
	require.NoError(t, err)
	assert.Greater(t, size, int64(0))

	// Verify mock expectations
	mockServerRepo.AssertExpectations(t)
}

func TestExportService_FormatBytes(t *testing.T) {
	service := &ExportService{}

	tests := []struct {
		bytes    uint64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1099511627776, "1.0 TiB"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d bytes", tt.bytes), func(t *testing.T) {
			result := service.formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringReadCloser(t *testing.T) {
	testData := "Hello, World!"
	rc := &StringReadCloser{Data: testData}

	// Test reading
	buf := make([]byte, 13)
	n, err := rc.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, testData, string(buf))

	// Test EOF
	n, err = rc.Read(buf)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)

	// Test close
	err = rc.Close()
	assert.NoError(t, err)
	assert.Equal(t, 0, rc.pos)
}

func TestExportServiceFactory(t *testing.T) {
	// Setup mocks
	mockServerRepo := &MockServerRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockRepo := &MockRepository{
		Server:  mockServerRepo,
		History: mockHistoryRepo,
		Cache:   mockCacheRepo,
	}
	mockLogger := &MockLogger{}

	// Create factory
	factory := NewExportServiceFactory(mockRepo, mockLogger)

	// Setup mock expectations for accessors
	mockRepo.On("ServerRepository").Return(mockServerRepo)
	mockRepo.On("HistoryRepository").Return(mockHistoryRepo)
	mockRepo.On("CacheRepository").Return(mockCacheRepo)

	// Create export service
	service := factory.CreateExportService()

	// Assertions
	assert.NotNil(t, service)
	assert.Equal(t, mockServerRepo, service.serverRepo)
	assert.Equal(t, mockHistoryRepo, service.historyRepo)
	assert.Equal(t, mockCacheRepo, service.cacheRepo)
	assert.Equal(t, mockLogger, service.logger)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// Mock repository for factory testing
type MockRepository struct {
	mock.Mock
	Server  repository.ServerRepository
	History repository.HistoryRepository
	Cache   repository.CacheRepository
}

func (m *MockRepository) ServerRepository() repository.ServerRepository {
	args := m.Called()
	if m.Server != nil {
		return m.Server
	}
	return args.Get(0).(repository.ServerRepository)
}

func (m *MockRepository) HistoryRepository() repository.HistoryRepository {
	args := m.Called()
	if m.History != nil {
		return m.History
	}
	return args.Get(0).(repository.HistoryRepository)
}

func (m *MockRepository) CacheRepository() repository.CacheRepository {
	args := m.Called()
	if m.Cache != nil {
		return m.Cache
	}
	return args.Get(0).(repository.CacheRepository)
}
