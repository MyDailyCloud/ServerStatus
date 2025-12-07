package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kanshan/ServerStatus/data-server/internal/config"
	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// RedisRepositoryFactory Redis仓库工厂
type RedisRepositoryFactory struct {
	client *redis.Client
	logger logger.Logger
	prefix string
}

// NewRedisRepositoryFactory 创建Redis仓库工厂
func NewRedisRepositoryFactory(cfg *config.CacheConfig, logger logger.Logger) (repository.RepositoryFactory, error) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  parseDuration(cfg.DialTimeout, 5*time.Second),
		ReadTimeout:  parseDuration(cfg.ReadTimeout, 3*time.Second),
		WriteTimeout: parseDuration(cfg.WriteTimeout, 3*time.Second),
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established successfully")

	return &RedisRepositoryFactory{
		client: client,
		logger: logger,
		prefix: "serverstatus",
	}, nil
}

// parseDuration 解析时间间隔
func parseDuration(str string, defaultDuration time.Duration) time.Duration {
	if str == "" {
		return defaultDuration
	}

	if duration, err := time.ParseDuration(str); err == nil {
		return duration
	}

	return defaultDuration
}

// NewRepository 创建完整的仓库实例
func (f *RedisRepositoryFactory) NewRepository(config interface{}) (repository.Repository, error) {
	return &RedisRepository{
		client:              f.client,
		logger:              f.logger,
		prefix:              f.prefix,
		cacheRepository:     NewRedisCacheRepository(f.client, f.logger, f.prefix),
		accessKeyRepository: NewRedisAccessKeyRepository(f.client, f.logger, f.prefix),
	}, nil
}

func (f *RedisRepositoryFactory) SupportedTypes() []string {
	return []string{"redis", "cache"}
}

// Close 关闭工厂和Redis连接
func (f *RedisRepositoryFactory) Close() error {
	if f.client != nil {
		return f.client.Close()
	}
	return nil
}

// RedisRepository 完整的Redis仓库实现
type RedisRepository struct {
	client              *redis.Client
	logger              logger.Logger
	prefix              string
	cacheRepository     repository.CacheRepository
	accessKeyRepository repository.AccessKeyRepository
}

// 以下实现满足 Repository 接口要求，但 Redis 仅提供缓存/访问密钥功能，其余返回未实现错误
var errNotImplemented = fmt.Errorf("not implemented in redis repository")

// ServerRepository methods
func (r *RedisRepository) CreateServer(ctx context.Context, server *models.ServerInfo) error {
	return errNotImplemented
}
func (r *RedisRepository) GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error) {
	return nil, errNotImplemented
}
func (r *RedisRepository) UpdateServer(ctx context.Context, server *models.ServerInfo) error {
	return errNotImplemented
}
func (r *RedisRepository) DeleteServer(ctx context.Context, sessionID string) error {
	return errNotImplemented
}
func (r *RedisRepository) GetAllServers(ctx context.Context, projectKey string, offset, limit int) ([]*models.ServerInfo, error) {
	return nil, errNotImplemented
}
func (r *RedisRepository) GetServersByHostname(ctx context.Context, hostname string) ([]*models.ServerInfo, error) {
	return nil, errNotImplemented
}
func (r *RedisRepository) GetServersByProject(ctx context.Context, projectKey string, pagination *repository.Pagination) ([]*models.ServerInfo, error) {
	return nil, errNotImplemented
}
func (r *RedisRepository) GetServerCount(ctx context.Context, projectKey string) (int, error) {
	return 0, errNotImplemented
}
func (r *RedisRepository) UpdateLastSeen(ctx context.Context, sessionID string) error {
	return errNotImplemented
}
func (r *RedisRepository) GetOnlineServers(ctx context.Context, projectKey string, timeout time.Duration) ([]*models.ServerInfo, error) {
	return nil, errNotImplemented
}

// HistoryRepository methods
func (r *RedisRepository) SaveHistoryData(ctx context.Context, data *models.SystemInfo) error {
	return errNotImplemented
}
func (r *RedisRepository) GetHostHistory(ctx context.Context, hostname, projectKey string, limit int) ([]*models.HistoryData, error) {
	return nil, errNotImplemented
}
func (r *RedisRepository) GetHistoryByTimeRange(ctx context.Context, hostname, projectKey string, start, end time.Time) ([]*models.HistoryData, error) {
	return nil, errNotImplemented
}
func (r *RedisRepository) GetHistoryByTimeRangePaged(ctx context.Context, hostname, projectKey string, start, end time.Time, pagination *repository.Pagination) ([]*models.HistoryData, int, error) {
	return nil, 0, errNotImplemented
}
func (r *RedisRepository) CleanupOldData(ctx context.Context, before time.Time) error {
	return errNotImplemented
}
func (r *RedisRepository) GetHistoryCount(ctx context.Context, hostname, projectKey string) (int, error) {
	return 0, errNotImplemented
}
func (r *RedisRepository) GetAggregatedData(ctx context.Context, hostname, projectKey string, interval time.Duration, limit int) ([]*models.HistoryData, error) {
	return nil, errNotImplemented
}

// CacheRepository methods (delegate to cacheRepository)
func (r *RedisRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.cacheRepository.Set(ctx, key, value, ttl)
}
func (r *RedisRepository) Get(ctx context.Context, key string, dest interface{}) error {
	return r.cacheRepository.Get(ctx, key, dest)
}
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	return r.cacheRepository.Delete(ctx, key)
}
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	return r.cacheRepository.Exists(ctx, key)
}
func (r *RedisRepository) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	return r.cacheRepository.SetMultiple(ctx, items, ttl)
}
func (r *RedisRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	return r.cacheRepository.GetMultiple(ctx, keys)
}
func (r *RedisRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	return r.cacheRepository.DeleteMultiple(ctx, keys)
}
func (r *RedisRepository) ClearPattern(ctx context.Context, pattern string) error {
	return r.cacheRepository.ClearPattern(ctx, pattern)
}
func (r *RedisRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.cacheRepository.Keys(ctx, pattern)
}
func (r *RedisRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return r.cacheRepository.GetStats(ctx)
}
func (r *RedisRepository) IsConnected() bool {
	return r.cacheRepository.IsConnected()
}
func (r *RedisRepository) GetType() string {
	return r.cacheRepository.GetType()
}

// AccessKey delegation
func (r *RedisRepository) SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error {
	return r.accessKeyRepository.SaveAccessKey(ctx, cacheKey, accessKey)
}
func (r *RedisRepository) GetAccessKey(ctx context.Context, accessKey string) (string, error) {
	return r.accessKeyRepository.GetAccessKey(ctx, accessKey)
}
func (r *RedisRepository) DeleteAccessKey(ctx context.Context, accessKey string) error {
	return r.accessKeyRepository.DeleteAccessKey(ctx, accessKey)
}
func (r *RedisRepository) GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error) {
	return r.accessKeyRepository.GenerateAccessKey(ctx, serverKey, projectKey)
}
func (r *RedisRepository) ValidateAccessKey(ctx context.Context, accessKey string) (string, error) {
	return r.accessKeyRepository.ValidateAccessKey(ctx, accessKey)
}
func (r *RedisRepository) CleanupExpiredKeys(ctx context.Context, ttl time.Duration) error {
	return r.accessKeyRepository.CleanupExpiredKeys(ctx, ttl)
}

// Ping 检查Redis连接
func (r *RedisRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err()
}

// Close 关闭仓库
func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// HealthChecker 健康检查器
func (r *RedisRepository) HealthChecker() repository.HealthChecker {
	cacheRepo := r.cacheRepository.(*RedisCacheRepository)
	return cacheRepo.HealthChecker()
}

// GetClient 获取Redis客户端（用于高级操作）
func (r *RedisRepository) GetClient() *redis.Client {
	return r.client
}

// GetPrefix 获取键前缀
func (r *RedisRepository) GetPrefix() string {
	return r.prefix
}

// Redis连接管理器
type RedisConnectionManager struct {
	client *redis.Client
	logger logger.Logger
	config *config.CacheConfig
}

// NewRedisConnectionManager 创建Redis连接管理器
func NewRedisConnectionManager(cfg *config.CacheConfig, logger logger.Logger) *RedisConnectionManager {
	return &RedisConnectionManager{
		config: cfg,
		logger: logger,
	}
}

// Connect 建立Redis连接
func (m *RedisConnectionManager) Connect() error {
	m.client = redis.NewClient(&redis.Options{
		Addr:         m.config.Address,
		Password:     m.config.Password,
		DB:           m.config.DB,
		PoolSize:     m.config.PoolSize,
		MinIdleConns: m.config.MinIdleConns,
		DialTimeout:  parseDuration(m.config.DialTimeout, 5*time.Second),
		ReadTimeout:  parseDuration(m.config.ReadTimeout, 3*time.Second),
		WriteTimeout: parseDuration(m.config.WriteTimeout, 3*time.Second),
		MaxRetries:   3,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	m.logger.Info("Redis connection established successfully")
	return nil
}

// Disconnect 断开Redis连接
func (m *RedisConnectionManager) Disconnect() error {
	if m.client != nil {
		err := m.client.Close()
		m.client = nil
		return err
	}
	return nil
}

// Reconnect 重新连接Redis
func (m *RedisConnectionManager) Reconnect() error {
	if err := m.Disconnect(); err != nil {
		m.logger.WithError(err).Warn("Failed to disconnect before reconnecting")
	}

	return m.Connect()
}

// IsConnected 检查连接状态
func (m *RedisConnectionManager) IsConnected() bool {
	if m.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.client.Ping(ctx).Err() == nil
}

// GetClient 获取Redis客户端
func (m *RedisConnectionManager) GetClient() *redis.Client {
	return m.client
}

// WatchHealth 监控Redis健康状态
func (m *RedisConnectionManager) WatchHealth(ctx context.Context, interval time.Duration) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !m.IsConnected() {
					select {
					case errCh <- fmt.Errorf("Redis connection lost"):
					default:
						// 避免阻塞
					}
					return
				}
			}
		}
	}()

	return errCh
}

// RedisPoolStats Redis连接池统计
type RedisPoolStats struct {
	Hits       uint64 `json:"hits"`
	Misses     uint64 `json:"misses"`
	Timeouts   uint64 `json:"timeouts"`
	TotalConns uint64 `json:"total_conns"`
	IdleConns  uint64 `json:"idle_conns"`
	StaleConns uint64 `json:"stale_conns"`
}

// GetPoolStats 获取连接池统计信息
func (m *RedisConnectionManager) GetPoolStats() RedisPoolStats {
	if m.client == nil {
		return RedisPoolStats{}
	}

	stats := m.client.PoolStats()
	return RedisPoolStats{
		Hits:       uint64(stats.Hits),
		Misses:     uint64(stats.Misses),
		Timeouts:   uint64(stats.Timeouts),
		TotalConns: uint64(stats.TotalConns),
		IdleConns:  uint64(stats.IdleConns),
		StaleConns: uint64(stats.StaleConns),
	}
}
