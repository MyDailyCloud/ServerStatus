package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/config"
	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepositoryFactory SQLite仓库工厂
type SQLiteRepositoryFactory struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteRepositoryFactory 创建SQLite仓库工厂
func NewSQLiteRepositoryFactory(cfg *config.DatabaseConfig, logger logger.Logger) (repository.RepositoryFactory, error) {
	// 初始化数据库连接
	db, err := InitializeDatabase(context.Background(), cfg.Path, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	if cfg.ConnMaxLifetime != "" {
		if d, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
			db.SetConnMaxLifetime(d)
		}
	}

	return &SQLiteRepositoryFactory{
		db:     db,
		logger: logger,
	}, nil
}

// NewRepository 创建完整的仓库实例
func (f *SQLiteRepositoryFactory) NewRepository(config interface{}) (repository.Repository, error) {
	return &SQLiteRepository{
		db:                  f.db,
		logger:              f.logger,
		serverRepository:    NewSQLiteServerRepository(f.db, f.logger),
		historyRepository:   NewSQLiteHistoryRepository(f.db, f.logger),
		cacheRepository:     NewSQLiteCacheRepository(f.db, f.logger),
		accessKeyRepository: NewSQLiteAccessKeyRepository(f.db, f.logger),
	}, nil
}

func (f *SQLiteRepositoryFactory) SupportedTypes() []string {
	return []string{"sqlite", "sqlite3"}
}

// Close 关闭工厂和数据库连接
func (f *SQLiteRepositoryFactory) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}

// SQLiteRepository 完整的SQLite仓库实现
type SQLiteRepository struct {
	db                  *sql.DB
	logger              logger.Logger
	serverRepository    repository.ServerRepository
	historyRepository   repository.HistoryRepository
	cacheRepository     repository.CacheRepository
	accessKeyRepository repository.AccessKeyRepository
}

// ServerRepository 返回服务器仓库
func (r *SQLiteRepository) ServerRepository() repository.ServerRepository {
	return r.serverRepository
}

// HistoryRepository 返回历史数据仓库
func (r *SQLiteRepository) HistoryRepository() repository.HistoryRepository {
	return r.historyRepository
}

// CacheRepository 返回缓存仓库
func (r *SQLiteRepository) CacheRepository() repository.CacheRepository {
	return r.cacheRepository
}

// AccessKeyRepository 返回访问密钥仓库
func (r *SQLiteRepository) AccessKeyRepository() repository.AccessKeyRepository {
	return r.accessKeyRepository
}

func (r *SQLiteRepository) CleanupExpiredKeys(ctx context.Context, ttl time.Duration) error {
	return r.accessKeyRepository.CleanupExpiredKeys(ctx, ttl)
}

func (r *SQLiteRepository) CleanupOldData(ctx context.Context, before time.Time) error {
	return r.historyRepository.CleanupOldData(ctx, before)
}

// Delegate CacheRepository methods
func (r *SQLiteRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.cacheRepository.Set(ctx, key, value, ttl)
}

func (r *SQLiteRepository) Get(ctx context.Context, key string, dest interface{}) error {
	return r.cacheRepository.Get(ctx, key, dest)
}

func (r *SQLiteRepository) Delete(ctx context.Context, key string) error {
	return r.cacheRepository.Delete(ctx, key)
}

func (r *SQLiteRepository) Exists(ctx context.Context, key string) (bool, error) {
	return r.cacheRepository.Exists(ctx, key)
}

func (r *SQLiteRepository) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	return r.cacheRepository.SetMultiple(ctx, items, ttl)
}

func (r *SQLiteRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	return r.cacheRepository.GetMultiple(ctx, keys)
}

func (r *SQLiteRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	return r.cacheRepository.DeleteMultiple(ctx, keys)
}

func (r *SQLiteRepository) ClearPattern(ctx context.Context, pattern string) error {
	return r.cacheRepository.ClearPattern(ctx, pattern)
}

func (r *SQLiteRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.cacheRepository.Keys(ctx, pattern)
}

func (r *SQLiteRepository) IsConnected() bool {
	return r.cacheRepository.IsConnected()
}

func (r *SQLiteRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return r.cacheRepository.GetStats(ctx)
}

func (r *SQLiteRepository) GetType() string {
	return r.cacheRepository.GetType()
}

// Delegate ServerRepository methods
func (r *SQLiteRepository) CreateServer(ctx context.Context, server *models.ServerInfo) error {
	return r.serverRepository.CreateServer(ctx, server)
}

func (r *SQLiteRepository) GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error) {
	return r.serverRepository.GetServer(ctx, sessionID)
}

func (r *SQLiteRepository) UpdateServer(ctx context.Context, server *models.ServerInfo) error {
	return r.serverRepository.UpdateServer(ctx, server)
}

func (r *SQLiteRepository) DeleteServer(ctx context.Context, sessionID string) error {
	return r.serverRepository.DeleteServer(ctx, sessionID)
}

func (r *SQLiteRepository) GetAllServers(ctx context.Context, projectKey string, offset, limit int) ([]*models.ServerInfo, error) {
	return r.serverRepository.GetAllServers(ctx, projectKey, offset, limit)
}

func (r *SQLiteRepository) GetServersByHostname(ctx context.Context, hostname string) ([]*models.ServerInfo, error) {
	return r.serverRepository.GetServersByHostname(ctx, hostname)
}

func (r *SQLiteRepository) GetServersByProject(ctx context.Context, projectKey string, pagination *repository.Pagination) ([]*models.ServerInfo, error) {
	return r.serverRepository.GetServersByProject(ctx, projectKey, pagination)
}

func (r *SQLiteRepository) GetServerCount(ctx context.Context, projectKey string) (int, error) {
	return r.serverRepository.GetServerCount(ctx, projectKey)
}

func (r *SQLiteRepository) UpdateLastSeen(ctx context.Context, sessionID string) error {
	return r.serverRepository.UpdateLastSeen(ctx, sessionID)
}

func (r *SQLiteRepository) GetOnlineServers(ctx context.Context, projectKey string, timeout time.Duration) ([]*models.ServerInfo, error) {
	return r.serverRepository.GetOnlineServers(ctx, projectKey, timeout)
}

// Delegate HistoryRepository methods
func (r *SQLiteRepository) SaveHistoryData(ctx context.Context, data *models.SystemInfo) error {
	return r.historyRepository.SaveHistoryData(ctx, data)
}

func (r *SQLiteRepository) GetHostHistory(ctx context.Context, hostname, projectKey string, limit int) ([]*models.HistoryData, error) {
	return r.historyRepository.GetHostHistory(ctx, hostname, projectKey, limit)
}

func (r *SQLiteRepository) GetHistoryByTimeRange(ctx context.Context, hostname, projectKey string, start, end time.Time) ([]*models.HistoryData, error) {
	return r.historyRepository.GetHistoryByTimeRange(ctx, hostname, projectKey, start, end)
}

func (r *SQLiteRepository) GetHistoryByTimeRangePaged(ctx context.Context, hostname, projectKey string, start, end time.Time, pagination *repository.Pagination) ([]*models.HistoryData, int, error) {
	return r.historyRepository.GetHistoryByTimeRangePaged(ctx, hostname, projectKey, start, end, pagination)
}

func (r *SQLiteRepository) GetHistoryCount(ctx context.Context, hostname, projectKey string) (int, error) {
	return r.historyRepository.GetHistoryCount(ctx, hostname, projectKey)
}

func (r *SQLiteRepository) GetAggregatedData(ctx context.Context, hostname, projectKey string, interval time.Duration, limit int) ([]*models.HistoryData, error) {
	return r.historyRepository.GetAggregatedData(ctx, hostname, projectKey, interval, limit)
}

// Delegate AccessKeyRepository methods
func (r *SQLiteRepository) SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error {
	return r.accessKeyRepository.SaveAccessKey(ctx, cacheKey, accessKey)
}

func (r *SQLiteRepository) GetAccessKey(ctx context.Context, accessKey string) (string, error) {
	return r.accessKeyRepository.GetAccessKey(ctx, accessKey)
}

func (r *SQLiteRepository) DeleteAccessKey(ctx context.Context, accessKey string) error {
	return r.accessKeyRepository.DeleteAccessKey(ctx, accessKey)
}

func (r *SQLiteRepository) GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error) {
	return r.accessKeyRepository.GenerateAccessKey(ctx, serverKey, projectKey)
}

func (r *SQLiteRepository) ValidateAccessKey(ctx context.Context, accessKey string) (string, error) {
	return r.accessKeyRepository.ValidateAccessKey(ctx, accessKey)
}

// Ping 检查数据库连接
func (r *SQLiteRepository) Ping() error {
	return r.db.Ping()
}

// Close 关闭仓库
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// HealthChecker 健康检查器
func (r *SQLiteRepository) HealthChecker() repository.HealthChecker {
	return &SQLiteHealthChecker{
		db:     r.db,
		logger: r.logger,
	}
}

// SQLiteHealthChecker SQLite健康检查器
type SQLiteHealthChecker struct {
	db     *sql.DB
	logger logger.Logger
}

// CheckHealth 检查健康状态
func (h *SQLiteHealthChecker) CheckHealth(ctx context.Context) error {
	if err := h.db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查表是否存在
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table'`
	var count int
	if err := h.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return fmt.Errorf("failed to check tables: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("no tables found in database")
	}

	return nil
}

// GetMetrics 获取数据库指标
func (h *SQLiteHealthChecker) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 获取数据库大小
	var size int64
	err := h.db.QueryRowContext(ctx, `SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()`).Scan(&size)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get database size")
	} else {
		metrics["database_size_bytes"] = size
		metrics["database_size_mb"] = float64(size) / (1024 * 1024)
	}

	// 获取连接池状态
	stats := h.db.Stats()
	metrics["open_connections"] = stats.OpenConnections
	metrics["in_use"] = stats.InUse
	metrics["idle"] = stats.Idle

	// 获取表记录数
	tables := []string{"servers", "server_history", "access_keys"}
	for _, table := range tables {
		var count int64
		err := h.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			h.logger.WithError(err).WithField("table", table).Warn("Failed to get table count")
			metrics[fmt.Sprintf("%s_count", table)] = 0
		} else {
			metrics[fmt.Sprintf("%s_count", table)] = count
		}
	}

	return metrics, nil
}
