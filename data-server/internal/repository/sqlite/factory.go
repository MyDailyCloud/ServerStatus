package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kanshan/ServerStatus/data-server/internal/config"
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

	return &SQLiteRepositoryFactory{
		db:     db,
		logger: logger,
	}, nil
}

// NewRepository 创建完整的仓库实例
func (f *SQLiteRepositoryFactory) NewRepository() (repository.Repository, error) {
	return &SQLiteRepository{
		db:                  f.db,
		logger:              f.logger,
		serverRepository:    NewSQLiteServerRepository(f.db, f.logger),
		historyRepository:   NewSQLiteHistoryRepository(f.db, f.logger),
		cacheRepository:     NewSQLiteCacheRepository(f.db, f.logger),
		accessKeyRepository: NewSQLiteAccessKeyRepository(f.db, f.logger),
	}, nil
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
