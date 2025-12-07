package repository

import (
	"context"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
)

// Repository 数据访问层接口
type Repository interface {
	ServerRepository
	HistoryRepository
	CacheRepository
	AccessKeyRepository
}

// ServerRepository 服务器数据访问接口
type ServerRepository interface {
	// 基础CRUD操作
	CreateServer(ctx context.Context, server *models.ServerInfo) error
	GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error)
	UpdateServer(ctx context.Context, server *models.ServerInfo) error
	DeleteServer(ctx context.Context, sessionID string) error

	// 查询操作
	GetAllServers(ctx context.Context, projectKey string, offset, limit int) ([]*models.ServerInfo, error)
	GetServersByHostname(ctx context.Context, hostname string) ([]*models.ServerInfo, error)
	GetServersByProject(ctx context.Context, projectKey string, pagination *Pagination) ([]*models.ServerInfo, error)
	GetServerCount(ctx context.Context, projectKey string) (int, error)

	// 状态操作
	UpdateLastSeen(ctx context.Context, sessionID string) error
	GetOnlineServers(ctx context.Context, projectKey string, timeout time.Duration) ([]*models.ServerInfo, error)

	// 健康检查
	Ping() error
	Close() error
}

// HistoryRepository 历史数据访问接口
type HistoryRepository interface {
	// 历史数据操作
	SaveHistoryData(ctx context.Context, data *models.SystemInfo) error
	GetHostHistory(ctx context.Context, hostname, projectKey string, limit int) ([]*models.HistoryData, error)
	GetHistoryByTimeRange(ctx context.Context, hostname, projectKey string, start, end time.Time) ([]*models.HistoryData, error)
	// 带分页的时间范围历史查询（返回数据与总数）
	GetHistoryByTimeRangePaged(ctx context.Context, hostname, projectKey string, start, end time.Time, pagination *Pagination) ([]*models.HistoryData, int, error)

	// 数据清理
	CleanupOldData(ctx context.Context, before time.Time) error
	GetHistoryCount(ctx context.Context, hostname, projectKey string) (int, error)

	// 聚合数据
	GetAggregatedData(ctx context.Context, hostname, projectKey string, interval time.Duration, limit int) ([]*models.HistoryData, error)

	// 健康检查
	Ping() error
}

// CacheRepository 缓存访问接口
type CacheRepository interface {
	// 基础缓存操作
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeleteMultiple(ctx context.Context, keys []string) error

	// 模式操作
	ClearPattern(ctx context.Context, pattern string) error
	Keys(ctx context.Context, pattern string) ([]string, error)

	// 统计操作
	GetStats(ctx context.Context) (map[string]interface{}, error)

	// 连接状态
	IsConnected() bool
	GetType() string // redis, memory
}

// AccessKeyRepository 访问密钥访问接口
type AccessKeyRepository interface {
	// 访问密钥操作
	SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error
	GetAccessKey(ctx context.Context, accessKey string) (string, error)
	DeleteAccessKey(ctx context.Context, accessKey string) error

	// 生成操作
	GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error)
	ValidateAccessKey(ctx context.Context, accessKey string) (string, error) // 返回projectKey

	// 清理操作
	CleanupExpiredKeys(ctx context.Context, ttl time.Duration) error
}

// Transaction 事务接口
type Transaction interface {
	Commit() error
	Rollback() error
}

// RepositoryFactory 数据库工厂接口
type RepositoryFactory interface {
	NewRepository(config interface{}) (Repository, error)
	SupportedTypes() []string
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	CheckHealth(ctx context.Context) error
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
}

type ServerFilter struct {
	ProjectKey     string
	Hostname       string
	OS             string
	Status         string
	Tags           []string
	LastSeenAfter  *time.Time
	LastSeenBefore *time.Time
	Page           int
	PageSize       int
	Limit          int
}

type Pagination struct {
	Page     int
	PageSize int
	Offset   int
	Limit    int
}

type ProjectStats struct {
	TotalServers   int
	OnlineServers  int
	OfflineServers int
	AvgCPUUsage    float64
	AvgMemoryUsage float64
	AvgDiskUsage   float64
	LastUpdateTime time.Time
}
