package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database 数据库结构体
type Database struct {
	db *sql.DB
}

// NewDatabase 创建新的数据库实例
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}

	// 创建表结构
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

// createTables 创建数据库表
func (d *Database) createTables() error {
	// 创建服务器信息表
	serverInfoSQL := `
	CREATE TABLE IF NOT EXISTS server_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT UNIQUE NOT NULL,
		hostname TEXT NOT NULL,
		project_key TEXT NOT NULL,
		latest_data TEXT,
		last_seen DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建历史数据表
	historyDataSQL := `
	CREATE TABLE IF NOT EXISTS history_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		hostname TEXT NOT NULL,
		project_key TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建历史数据表索引
	historyDataIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_session_timestamp ON history_data(session_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_hostname_timestamp ON history_data(hostname, timestamp);
	CREATE INDEX IF NOT EXISTS idx_project_key ON history_data(project_key);`

	// 创建访问密钥缓存表
	accessKeySQL := `
	CREATE TABLE IF NOT EXISTS access_key_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cache_key TEXT UNIQUE NOT NULL,
		access_key TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建访问密钥缓存表索引
	accessKeyIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_cache_key ON access_key_cache(cache_key);`

	// 执行创建表语句
	if _, err := d.db.Exec(serverInfoSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(historyDataSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(accessKeySQL); err != nil {
		return err
	}

	// 执行创建索引语句
	if _, err := d.db.Exec(historyDataIndexSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(accessKeyIndexSQL); err != nil {
		return err
	}

	return nil
}

// SaveServerInfo 保存服务器信息
func (d *Database) SaveServerInfo(sessionID, hostname, projectKey string, latestData *SystemInfo) error {
	// 序列化最新数据
	dataJSON, err := json.Marshal(latestData)
	if err != nil {
		return fmt.Errorf("序列化服务器信息失败: %w", err)
	}

	// 使用UPSERT语法（SQLite 3.24+）
	query := `
	INSERT INTO server_info (session_id, hostname, project_key, latest_data, last_seen)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(session_id)
	DO UPDATE SET
		hostname = excluded.hostname,
		project_key = excluded.project_key,
		latest_data = excluded.latest_data,
		last_seen = excluded.last_seen,
		updated_at = CURRENT_TIMESTAMP;
	`

	_, err = d.db.Exec(query, sessionID, hostname, projectKey, string(dataJSON), time.Now())
	if err != nil {
		return fmt.Errorf("保存服务器信息失败: %w", err)
	}
	return nil
}

// SaveHistoryData 保存历史数据
func (d *Database) SaveHistoryData(sessionID, hostname, projectKey string, data *SystemInfo) error {
	// 序列化数据
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化历史数据失败: %w", err)
	}

	query := `
	INSERT INTO history_data (session_id, hostname, project_key, timestamp, data)
	VALUES (?, ?, ?, ?, ?);
	`

	_, err = d.db.Exec(query, sessionID, hostname, projectKey, data.Timestamp, string(dataJSON))
	if err != nil {
		return fmt.Errorf("保存历史数据失败: %w", err)
	}
	return nil
}

// GetServerInfo 获取服务器信息
func (d *Database) GetServerInfo(sessionID string) (*ServerInfo, error) {
	var latestData string
	var lastSeen time.Time

	query := `
	SELECT latest_data, last_seen
	FROM server_info
	WHERE session_id = ?;
	`

	err := d.db.QueryRow(query, sessionID).Scan(&latestData, &lastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 反序列化最新数据
	var systemInfo SystemInfo
	if err := json.Unmarshal([]byte(latestData), &systemInfo); err != nil {
		return nil, err
	}

	serverInfo := &ServerInfo{
		Latest:   &systemInfo,
		LastSeen: lastSeen,
		History:  make([]*SystemInfo, 0),
	}

	return serverInfo, nil
}

// GetHistoryData 获取历史数据
func (d *Database) GetHistoryData(sessionID string, limit int) ([]*SystemInfo, error) {
	query := `
	SELECT data, timestamp
	FROM history_data
	WHERE session_id = ?
	ORDER BY timestamp DESC
	LIMIT ?;
	`

	rows, err := d.db.Query(query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*SystemInfo
	for rows.Next() {
		var dataStr string
		var timestamp time.Time

		if err := rows.Scan(&dataStr, &timestamp); err != nil {
			continue
		}

		var systemInfo SystemInfo
		if err := json.Unmarshal([]byte(dataStr), &systemInfo); err != nil {
			continue
		}

		history = append(history, &systemInfo)
	}

	return history, nil
}

// GetAllServers 获取所有服务器（带分页支持）
func (d *Database) GetAllServers(projectKey string, offset, limit int) ([]*ServerInfo, error) {
	var query string
	var args []interface{}

	if projectKey == "public" || projectKey == "" {
		query = `
		SELECT si.session_id, si.hostname, si.latest_data, si.last_seen
		FROM server_info si
		ORDER BY si.hostname
		LIMIT ? OFFSET ?;
		`
		args = []interface{}{limit, offset}
	} else {
		query = `
		SELECT si.session_id, si.hostname, si.latest_data, si.last_seen
		FROM server_info si
		WHERE si.project_key = ?
		ORDER BY si.hostname
		LIMIT ? OFFSET ?;
		`
		args = []interface{}{projectKey, limit, offset}
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*ServerInfo
	for rows.Next() {
		var sessionID, hostname, latestDataStr string
		var lastSeen time.Time

		if err := rows.Scan(&sessionID, &hostname, &latestDataStr, &lastSeen); err != nil {
			continue
		}

		var latestData SystemInfo
		if err := json.Unmarshal([]byte(latestDataStr), &latestData); err != nil {
			continue
		}

		serverInfo := &ServerInfo{
			Latest:   &latestData,
			LastSeen: lastSeen,
			History:  make([]*SystemInfo, 0),
		}

		servers = append(servers, serverInfo)
	}

	return servers, nil
}

// GetServerCount 获取服务器总数
func (d *Database) GetServerCount(projectKey string) (int, error) {
	var query string
	var count int

	if projectKey == "public" || projectKey == "" {
		query = "SELECT COUNT(*) FROM server_info;"
		err := d.db.QueryRow(query).Scan(&count)
		return count, err
	} else {
		query = "SELECT COUNT(*) FROM server_info WHERE project_key = ?;"
		err := d.db.QueryRow(query, projectKey).Scan(&count)
		return count, err
	}
}

// CleanupOldData 清理旧数据
func (d *Database) CleanupOldData(olderThan time.Duration, maxHistoryPerServer int) error {
	// 清理超过指定时间的历史数据
	cutoffTime := time.Now().Add(-olderThan)

	query := `
	DELETE FROM history_data
	WHERE timestamp < ?;
	`

	if _, err := d.db.Exec(query, cutoffTime); err != nil {
		return err
	}

	// 为每个服务器保留最多N条最新记录
	query = `
	DELETE FROM history_data
	WHERE id NOT IN (
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY timestamp DESC) as rn
			FROM history_data
		) t WHERE rn <= ?
	);
	`

	if _, err := d.db.Exec(query, maxHistoryPerServer); err != nil {
		return err
	}

	return nil
}

// SaveAccessKeyCache 保存访问密钥缓存
func (d *Database) SaveAccessKeyCache(cacheKey, accessKey string) error {
	query := `
	INSERT OR REPLACE INTO access_key_cache (cache_key, access_key)
	VALUES (?, ?);
	`

	_, err := d.db.Exec(query, cacheKey, accessKey)
	return err
}

// GetAccessKeyCache 获取访问密钥缓存
func (d *Database) GetAccessKeyCache(cacheKey string) (string, error) {
	var accessKey string

	query := `
	SELECT access_key
	FROM access_key_cache
	WHERE cache_key = ?;
	`

	err := d.db.QueryRow(query, cacheKey).Scan(&accessKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return accessKey, nil
}

// GetUUIDStats 获取UUID统计信息
func (d *Database) GetUUIDStats() (map[string]interface{}, error) {
	query := `
	SELECT
		COUNT(*) as total_servers,
		COUNT(CASE WHEN session_id != hostname THEN 1 END) as active_uuids,
		COUNT(CASE WHEN session_id = hostname THEN 1 END) as hostname_only
	FROM server_info;
	`

	var totalServers, activeUUIDs, hostnameOnly int
	err := d.db.QueryRow(query).Scan(&totalServers, &activeUUIDs, &hostnameOnly)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{
		"total_servers": totalServers,
		"active_uuids":  activeUUIDs,
		"hostname_only": hostnameOnly,
		"timestamp":     time.Now(),
		"description":   "使用我们服务的设备统计",
	}

	return response, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping 检查数据库连接
func (d *Database) Ping() error {
	return d.db.Ping()
}