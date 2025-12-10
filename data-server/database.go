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

	// 创建访客事件表
	visitorEventsSQL := `
	CREATE TABLE IF NOT EXISTS visitor_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_key TEXT NOT NULL,
		domain TEXT,
		page_url TEXT,
		referrer TEXT,
		user_agent TEXT,
		ip TEXT,
		session_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建访问密钥缓存表索引
	accessKeyIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_cache_key ON access_key_cache(cache_key);`

	// 创建访客事件索引
	visitorEventIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_visitor_project_time ON visitor_events(project_key, created_at);
	CREATE INDEX IF NOT EXISTS idx_visitor_page ON visitor_events(page_url);
	CREATE INDEX IF NOT EXISTS idx_visitor_domain ON visitor_events(domain);`

	// 创建域名绑定表
	domainBindingSQL := `
	CREATE TABLE IF NOT EXISTS domain_bindings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT UNIQUE NOT NULL,
		project_key TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建用户表
	userSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		github_id TEXT UNIQUE NOT NULL,
		login TEXT,
		name TEXT,
		avatar_url TEXT,
		email TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建会话表
	sessionSQL := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		session_token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// 创建用户配置表
	userConfigSQL := `
	CREATE TABLE IF NOT EXISTS user_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		config TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// 用户/会话索引
	userIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	`

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

	if _, err := d.db.Exec(visitorEventsSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(domainBindingSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(userSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(sessionSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(userConfigSQL); err != nil {
		return err
	}

	// 执行创建索引语句
	if _, err := d.db.Exec(historyDataIndexSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(accessKeyIndexSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(visitorEventIndexSQL); err != nil {
		return err
	}

	if _, err := d.db.Exec(userIndexSQL); err != nil {
		return err
	}

	// 迁移：为旧表添加domain列（忽略已存在错误）
	_, _ = d.db.Exec("ALTER TABLE visitor_events ADD COLUMN domain TEXT;")

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

// GetHistoryByTimeRange 按时间范围获取历史数据（倒序）
func (d *Database) GetHistoryByTimeRange(sessionID, projectKey string, from, to time.Time, limit int) ([]*SystemInfo, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id 不能为空")
	}
	if from.IsZero() {
		from = time.Now().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now()
	}
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}

	where := "session_id = ? AND timestamp BETWEEN ? AND ?"
	args := []interface{}{sessionID, from, to, limit}
	if projectKey != "" {
		where = "session_id = ? AND project_key = ? AND timestamp BETWEEN ? AND ?"
		args = []interface{}{sessionID, projectKey, from, to, limit}
	}

	query := fmt.Sprintf(`
	SELECT data, timestamp
	FROM history_data
	WHERE %s
	ORDER BY timestamp DESC
	LIMIT ?;
	`, where)

	rows, err := d.db.Query(query, args...)
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

// SaveVisitorEvent 保存访客事件
func (d *Database) SaveVisitorEvent(event *VisitorEvent) error {
	if event == nil {
		return fmt.Errorf("访客事件为空")
	}

	query := `
	INSERT INTO visitor_events (project_key, domain, page_url, referrer, user_agent, ip, session_id, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`

	_, err := d.db.Exec(query, event.ProjectKey, event.Domain, event.PageURL, event.Referrer, event.UserAgent, event.IP, event.SessionID, event.Timestamp)
	return err
}

// GetVisitorStats 获取访客统计信息
func (d *Database) GetVisitorStats(projectKey string, since time.Time) (*VisitorStats, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("统计时间范围无效")
	}

	where := "created_at >= ?"
	args := []interface{}{since}
	if projectKey != "" {
		where += " AND project_key = ?"
		args = append(args, projectKey)
	}

	// 总访问量
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM visitor_events WHERE %s", where)
	var total int
	if err := d.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// 独立IP数
	uniqueIPQuery := fmt.Sprintf("SELECT COUNT(DISTINCT ip) FROM visitor_events WHERE %s", where)
	var uniqueIPs int
	if err := d.db.QueryRow(uniqueIPQuery, args...).Scan(&uniqueIPs); err != nil {
		return nil, err
	}

	// 每日访问趋势
	dailyQuery := fmt.Sprintf(`
	SELECT strftime('%%Y-%%m-%%d', created_at) AS day, COUNT(*) AS cnt
	FROM visitor_events
	WHERE %s
	GROUP BY day
	ORDER BY day;
	`, where)
	rows, err := d.db.Query(dailyQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daily []DailyVisitorStat
	for rows.Next() {
		var stat DailyVisitorStat
		if err := rows.Scan(&stat.Date, &stat.Count); err != nil {
			continue
		}
		daily = append(daily, stat)
	}

	// 最受欢迎的页面
	pageQuery := fmt.Sprintf(`
	SELECT page_url, COUNT(*) AS cnt
	FROM visitor_events
	WHERE %s AND page_url != ''
	GROUP BY page_url
	ORDER BY cnt DESC
	LIMIT 10;
	`, where)
	pageRows, err := d.db.Query(pageQuery, args...)
	if err != nil {
		return nil, err
	}
	defer pageRows.Close()

	var topPages []PageVisitorStat
	for pageRows.Next() {
		var stat PageVisitorStat
		if err := pageRows.Scan(&stat.Page, &stat.Count); err != nil {
			continue
		}
		topPages = append(topPages, stat)
	}

	stats := &VisitorStats{
		ProjectKey:  projectKey,
		From:        since,
		To:          time.Now(),
		TotalVisits: total,
		UniqueIPs:   uniqueIPs,
		Daily:       daily,
		TopPages:    topPages,
	}

	return stats, nil
}

// AggregatedVisitorItem 访客聚合项
type AggregatedVisitorItem struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// User 用户
type User struct {
	ID        int64
	GithubID  string
	Login     string
	Name      string
	AvatarURL string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Session 会话
type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

// GetVisitorAggregation 通用分组聚合
func (d *Database) GetVisitorAggregation(projectKey, groupBy string, since time.Time, limit int) ([]AggregatedVisitorItem, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("统计时间范围无效")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var column string
	switch groupBy {
	case "page":
		column = "page_url"
	case "referrer":
		column = "referrer"
	case "domain":
		column = "domain"
	case "ua", "user_agent":
		column = "user_agent"
	default:
		return nil, fmt.Errorf("不支持的分组字段: %s", groupBy)
	}

	where := "created_at >= ?"
	args := []interface{}{since}
	if projectKey != "" {
		where += " AND project_key = ?"
		args = append(args, projectKey)
	}

	query := fmt.Sprintf(`
	SELECT %s as key, COUNT(*) as cnt
	FROM visitor_events
	WHERE %s AND %s != ''
	GROUP BY %s
	ORDER BY cnt DESC
	LIMIT ?;
	`, column, where, column, column)

	args = append(args, limit)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AggregatedVisitorItem
	for rows.Next() {
		var item AggregatedVisitorItem
		if err := rows.Scan(&item.Key, &item.Count); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// DomainBinding 域名绑定
type DomainBinding struct {
	Domain     string `json:"domain"`
	ProjectKey string `json:"project_key"`
}

// UpsertDomainBinding 新增或更新域名绑定
func (d *Database) UpsertDomainBinding(binding DomainBinding) error {
	if binding.Domain == "" || binding.ProjectKey == "" {
		return fmt.Errorf("域名或项目密钥不能为空")
	}

	query := `
	INSERT INTO domain_bindings (domain, project_key)
	VALUES (?, ?)
	ON CONFLICT(domain) DO UPDATE SET project_key=excluded.project_key;
	`
	_, err := d.db.Exec(query, binding.Domain, binding.ProjectKey)
	return err
}

// GetProjectKeyByDomain 根据域名获取项目密钥
func (d *Database) GetProjectKeyByDomain(domain string) (string, error) {
	if domain == "" {
		return "", nil
	}
	query := `
	SELECT project_key FROM domain_bindings WHERE domain = ? LIMIT 1;
	`
	var projectKey string
	err := d.db.QueryRow(query, domain).Scan(&projectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return projectKey, nil
}

// ListDomainBindings 列出域名绑定
func (d *Database) ListDomainBindings() ([]DomainBinding, error) {
	rows, err := d.db.Query(`SELECT domain, project_key FROM domain_bindings ORDER BY domain;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DomainBinding
	for rows.Next() {
		var b DomainBinding
		if err := rows.Scan(&b.Domain, &b.ProjectKey); err != nil {
			continue
		}
		list = append(list, b)
	}
	return list, nil
}

// SaveOrUpdateUser 插入或更新用户，返回用户ID
func (d *Database) SaveOrUpdateUser(user *User) (int64, error) {
	if user == nil || user.GithubID == "" {
		return 0, fmt.Errorf("用户信息无效")
	}

	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 尝试更新
	update := `
	UPDATE users
	SET login = ?, name = ?, avatar_url = ?, email = ?, updated_at = CURRENT_TIMESTAMP
	WHERE github_id = ?;
	`
	if _, err := tx.Exec(update, user.Login, user.Name, user.AvatarURL, user.Email, user.GithubID); err != nil {
		return 0, err
	}

	// 获取ID
	var id int64
	err = tx.QueryRow(`SELECT id FROM users WHERE github_id = ?;`, user.GithubID).Scan(&id)
	if err == sql.ErrNoRows {
		insert := `
		INSERT INTO users (github_id, login, name, avatar_url, email)
		VALUES (?, ?, ?, ?, ?);
		`
		res, err := tx.Exec(insert, user.GithubID, user.Login, user.Name, user.AvatarURL, user.Email)
		if err != nil {
			return 0, err
		}
		id, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// CreateSession 创建会话
func (d *Database) CreateSession(userID int64, token string, expiresAt time.Time) error {
	query := `
	INSERT INTO sessions (user_id, session_token, expires_at)
	VALUES (?, ?, ?);
	`
	_, err := d.db.Exec(query, userID, token, expiresAt)
	return err
}

// GetSession 根据token获取会话
func (d *Database) GetSession(token string) (*Session, error) {
	var s Session
	query := `
	SELECT session_token, user_id, expires_at, created_at
	FROM sessions
	WHERE session_token = ?;
	`
	err := d.db.QueryRow(query, token).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// DeleteSession 删除会话
func (d *Database) DeleteSession(token string) error {
	_, err := d.db.Exec(`DELETE FROM sessions WHERE session_token = ?;`, token)
	return err
}

// GetUserByID 根据ID获取用户
func (d *Database) GetUserByID(id int64) (*User, error) {
	var u User
	query := `
	SELECT id, github_id, login, name, avatar_url, email, created_at, updated_at
	FROM users
	WHERE id = ?;
	`
	err := d.db.QueryRow(query, id).Scan(&u.ID, &u.GithubID, &u.Login, &u.Name, &u.AvatarURL, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// UpsertUserConfig 创建或更新用户配置
func (d *Database) UpsertUserConfig(userID int64, config string) error {
	if config == "" {
		config = "{}"
	}
	query := `
	INSERT INTO user_configs (user_id, config)
	VALUES (?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		config = excluded.config,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err := d.db.Exec(query, userID, config)
	return err
}

// GetUserConfig 获取用户配置
func (d *Database) GetUserConfig(userID int64) (string, error) {
	var config string
	err := d.db.QueryRow(`SELECT config FROM user_configs WHERE user_id = ?;`, userID).Scan(&config)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return config, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping 检查数据库连接
func (d *Database) Ping() error {
	return d.db.Ping()
}
