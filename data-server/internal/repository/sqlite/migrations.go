package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Migration 数据库迁移
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// GetMigrations 获取所有迁移
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create servers table",
			SQL: `
				CREATE TABLE IF NOT EXISTS servers (
					session_id TEXT PRIMARY KEY,
					hostname TEXT NOT NULL,
					project_key TEXT NOT NULL DEFAULT '',
					os TEXT,
					arch TEXT,
					cpu_cores INTEGER,
					memory_total BIGINT,
					disk_total BIGINT,
					uptime BIGINT,
					boot_time DATETIME,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_servers_hostname ON servers(hostname);
				CREATE INDEX IF NOT EXISTS idx_servers_project_key ON servers(project_key);
				CREATE INDEX IF NOT EXISTS idx_servers_updated_at ON servers(updated_at);
			`,
		},
		{
			Version:     2,
			Description: "Create server_history table",
			SQL: `
				CREATE TABLE IF NOT EXISTS server_history (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					session_id TEXT NOT NULL,
					hostname TEXT NOT NULL,
					project_key TEXT NOT NULL DEFAULT '',
					cpu_usage REAL,
					memory_used BIGINT,
					memory_available BIGINT,
					disk_used BIGINT,
					disk_available BIGINT,
					network_rx BIGINT,
					network_tx BIGINT,
					load_avg REAL,
					process_count INTEGER,
					timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (session_id) REFERENCES servers(session_id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_server_history_session_id ON server_history(session_id);
				CREATE INDEX IF NOT EXISTS idx_server_history_hostname ON server_history(hostname);
				CREATE INDEX IF NOT EXISTS idx_server_history_timestamp ON server_history(timestamp);
				CREATE INDEX IF NOT EXISTS idx_server_history_project_key ON server_history(project_key);
			`,
		},
		{
			Version:     3,
			Description: "Create access_keys table",
			SQL: `
				CREATE TABLE IF NOT EXISTS access_keys (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					access_key TEXT UNIQUE NOT NULL,
					cache_key TEXT NOT NULL,
					project_key TEXT NOT NULL,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at DATETIME,
					is_active BOOLEAN NOT NULL DEFAULT 1
				);

				CREATE INDEX IF NOT EXISTS idx_access_keys_key ON access_keys(access_key);
				CREATE INDEX IF NOT EXISTS idx_access_keys_project_key ON access_keys(project_key);
				CREATE INDEX IF NOT EXISTS idx_access_keys_expires_at ON access_keys(expires_at);
			`,
		},
		{
			Version:     4,
			Description: "Create migration_history table",
			SQL: `
				CREATE TABLE IF NOT EXISTS migration_history (
					version INTEGER PRIMARY KEY,
					description TEXT NOT NULL,
					applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
		{
			Version:     5,
			Description: "Add server status fields",
			SQL: `
				ALTER TABLE servers ADD COLUMN is_online BOOLEAN DEFAULT 1;
				ALTER TABLE servers ADD COLUMN last_ping DATETIME;
			`,
		},
		{
			Version:     6,
			Description: "Create cache table",
			SQL: `
				CREATE TABLE IF NOT EXISTS cache (
					key TEXT PRIMARY KEY,
					value TEXT NOT NULL,
					expires_at DATETIME,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_cache_expires_at ON cache(expires_at);
				CREATE INDEX IF NOT EXISTS idx_cache_created_at ON cache(created_at);
			`,
		},
	}
}

// Migrator 数据库迁移器
type Migrator struct {
	db     *sql.DB
	logger logger.Logger
}

// NewMigrator 创建迁移器
func NewMigrator(db *sql.DB, logger logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// RunMigrations 运行数据库迁移
func (m *Migrator) RunMigrations(ctx context.Context) error {
	// 确保迁移历史表存在
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	// 获取当前版本
	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	m.logger.WithField("current_version", currentVersion).Info("Starting database migration")

	// 获取所有迁移
	migrations := GetMigrations()

	// 执行待执行的迁移
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		m.logger.WithFields(map[string]interface{}{
			"version":     migration.Version,
			"description": migration.Description,
		}).Info("Applying migration")

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		m.logger.WithFields(map[string]interface{}{
			"version":     migration.Version,
			"description": migration.Description,
		}).Info("Migration applied successfully")
	}

	m.logger.Info("Database migration completed successfully")
	return nil
}

// ensureMigrationTable 确保迁移表存在
func (m *Migrator) ensureMigrationTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS migration_history (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migration_history table: %w", err)
	}

	return nil
}

// getCurrentVersion 获取当前数据库版本
func (m *Migrator) getCurrentVersion(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM migration_history`

	var version int
	err := m.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// applyMigration 应用单个迁移
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	// 开始事务
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 执行迁移SQL
	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// 记录迁移历史
	insertQuery := `INSERT INTO migration_history (version, description) VALUES (?, ?)`
	if _, err := tx.ExecContext(ctx, insertQuery, migration.Version, migration.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// InitializeDatabase 初始化数据库
func InitializeDatabase(ctx context.Context, dbPath string, logger logger.Logger) (*sql.DB, error) {
	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 启用外键约束
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// 减少写入竞争时的错误，设置忙等待超时
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		logger.Warn("Failed to set busy_timeout, continuing without it")
	}

	// 启用WAL模式以提高并发性能
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		logger.Warn("Failed to enable WAL mode, continuing without it")
	}

	// 运行迁移
	migrator := NewMigrator(db, logger)
	if err := migrator.RunMigrations(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database initialized successfully")
	return db, nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
