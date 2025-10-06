package pool

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/config"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Manager 连接池管理器接口
type Manager interface {
	GetConnection() (*sql.DB, error)
	PutConnection(db *sql.DB) error
	Close() error
	GetStats() PoolStats
	HealthCheck(ctx context.Context) error
}

// PoolStats 连接池统计信息
type PoolStats struct {
	TotalConnections    int           `json:"total_connections"`
	ActiveConnections   int           `json:"active_connections"`
	IdleConnections     int           `json:"idle_connections"`
	WaitingConnections  int           `json:"waiting_connections"`
	MaxConnections      int           `json:"max_connections"`
	ConnectionLifetime  time.Duration `json:"connection_lifetime"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	CreatedAt           time.Time     `json:"created_at"`
	LastHealthCheck     time.Time     `json:"last_health_check"`
}

// ConnectionPool 连接池实现
type ConnectionPool struct {
	config         *config.DatabaseConfig
	logger         logger.Logger
	maxConnections int
	idleTimeout    time.Duration
	lifetime       time.Duration
	healthInterval time.Duration

	mu              sync.RWMutex
	connections     []*sql.DB
	activeConns     map[*sql.DB]bool
	waitingQueue    chan *ConnectionRequest
	stats           PoolStats
	healthTicker    *time.Ticker
	stopCh          chan struct{}
	closed          bool
	createdAt       time.Time
	lastHealthCheck time.Time
}

// ConnectionRequest 连接请求
type ConnectionRequest struct {
	Connection chan *sql.DB
	Error      chan error
	Timeout    time.Duration
}

// NewConnectionPool 创建连接池
func NewConnectionPool(cfg *config.DatabaseConfig, logger logger.Logger) Manager {
	return &ConnectionPool{
		config:         cfg,
		logger:         logger,
		maxConnections: cfg.MaxOpenConns,
		idleTimeout:    parseDuration(cfg.ConnMaxLifetime, 5*time.Minute),
		lifetime:       parseDuration(cfg.ConnMaxLifetime, 30*time.Minute),
		healthInterval: 30 * time.Second,
		connections:    make([]*sql.DB, 0, cfg.MaxOpenConns),
		activeConns:    make(map[*sql.DB]bool),
		waitingQueue:   make(chan *ConnectionRequest, 100),
		createdAt:      time.Now(),
		stopCh:         make(chan struct{}),
	}
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

// GetConnection 获取数据库连接
func (p *ConnectionPool) GetConnection() (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	// 首先尝试从空闲连接中获取
	for i, conn := range p.connections {
		if !p.activeConns[conn] {
			// 检查连接是否有效
			if p.isConnectionValid(conn) {
				p.activeConns[conn] = true
				p.updateStats()
				p.logger.Debug("Reusing existing connection from pool")
				return conn, nil
			} else {
				// 连接无效，从池中移除
				p.removeConnection(i)
				conn.Close()
			}
		}
	}

	// 如果没有空闲连接且还能创建新连接
	if len(p.connections) < p.maxConnections {
		conn, err := p.createConnection()
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}
		p.activeConns[conn] = true
		p.updateStats()
		p.logger.Debug("Created new connection")
		return conn, nil
	}

	// 连接池已满，需要等待
	return p.waitForConnection()
}

// PutConnection 归还连接到池中
func (p *ConnectionPool) PutConnection(conn *sql.DB) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("connection pool is closed")
	}

	// 检查连接是否属于这个池
	belongsToPool := false
	for _, poolConn := range p.connections {
		if poolConn == conn {
			belongsToPool = true
			break
		}
	}

	if !belongsToPool {
		return fmt.Errorf("connection does not belong to this pool")
	}

	// 标记为空闲
	p.activeConns[conn] = false
	p.updateStats()

	// 检查是否有等待的请求
	select {
	case req := <-p.waitingQueue:
		// 有等待的请求，直接分配
		p.activeConns[conn] = true
		select {
		case req.Connection <- conn:
			p.logger.Debug("Connection assigned to waiting request")
		case <-time.After(req.Timeout):
			// 请求超时，归还连接
			p.activeConns[conn] = false
			req.Error <- fmt.Errorf("connection request timeout")
		}
	default:
		// 没有等待的请求，连接保持空闲
		p.logger.Debug("Connection returned to pool")
	}

	return nil
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// 停止健康检查
	if p.healthTicker != nil {
		p.healthTicker.Stop()
	}

	// 关闭停止通道
	close(p.stopCh)

	// 关闭所有连接
	for _, conn := range p.connections {
		conn.Close()
	}

	// 清理资源
	p.connections = nil
	p.activeConns = nil

	p.logger.Info("Connection pool closed")
	return nil
}

// GetStats 获取连接池统计信息
func (p *ConnectionPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.stats
}

// HealthCheck 执行健康检查
func (p *ConnectionPool) HealthCheck(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("connection pool is closed")
	}

	// 检查至少有一个可用连接
	if len(p.connections) == 0 {
		return fmt.Errorf("no connections in pool")
	}

	// 测试一个连接
	for _, conn := range p.connections {
		if err := conn.PingContext(ctx); err != nil {
			p.logger.WithError(err).Warn("Connection health check failed")
			return fmt.Errorf("connection health check failed: %w", err)
		}
		break
	}

	p.lastHealthCheck = time.Now()
	p.updateStats()
	return nil
}

// createConnection 创建新连接
func (p *ConnectionPool) createConnection() (*sql.DB, error) {
	// 这里应该根据配置创建适当的数据库连接
	// 为了简化，我们返回一个模拟的连接
	return nil, fmt.Errorf("database connection creation not implemented")
}

// isConnectionValid 检查连接是否有效
func (p *ConnectionPool) isConnectionValid(conn *sql.DB) bool {
	if conn == nil {
		return false
	}

	// 检查连接年龄
	if time.Since(p.createdAt) > p.lifetime {
		return false
	}

	// 执行ping检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		p.logger.WithError(err).Debug("Connection ping failed")
		return false
	}

	return true
}

// removeConnection 从池中移除连接
func (p *ConnectionPool) removeConnection(index int) {
	if index >= 0 && index < len(p.connections) {
		conn := p.connections[index]
		delete(p.activeConns, conn)
		p.connections = append(p.connections[:index], p.connections[index+1:]...)
	}
}

// waitForConnection 等待可用连接
func (p *ConnectionPool) waitForConnection() (*sql.DB, error) {
	req := &ConnectionRequest{
		Connection: make(chan *sql.DB, 1),
		Error:      make(chan error, 1),
		Timeout:    30 * time.Second,
	}

	// 将请求加入等待队列
	select {
	case p.waitingQueue <- req:
		p.logger.Debug("Connection request added to waiting queue")
	default:
		return nil, fmt.Errorf("waiting queue is full")
	}

	// 等待连接或超时
	select {
	case conn := <-req.Connection:
		return conn, nil
	case err := <-req.Error:
		return nil, err
	case <-time.After(req.Timeout):
		return nil, fmt.Errorf("connection request timeout")
	}
}

// updateStats 更新统计信息
func (p *ConnectionPool) updateStats() {
	activeCount := 0
	for _, active := range p.activeConns {
		if active {
			activeCount++
		}
	}

	p.stats = PoolStats{
		TotalConnections:    len(p.connections),
		ActiveConnections:   activeCount,
		IdleConnections:     len(p.connections) - activeCount,
		WaitingConnections:  len(p.waitingQueue),
		MaxConnections:      p.maxConnections,
		ConnectionLifetime:  p.lifetime,
		IdleTimeout:         p.idleTimeout,
		HealthCheckInterval: p.healthInterval,
		CreatedAt:           p.createdAt,
		LastHealthCheck:     p.lastHealthCheck,
	}
}

// StartHealthCheck 启动健康检查
func (p *ConnectionPool) StartHealthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.healthTicker != nil {
		return // 已经启动
	}

	p.healthTicker = time.NewTicker(p.healthInterval)

	go func() {
		for {
			select {
			case <-p.stopCh:
				return
			case <-p.healthTicker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := p.HealthCheck(ctx); err != nil {
					p.logger.WithError(err).Error("Pool health check failed")
				}
				cancel()
			}
		}
	}()

	p.logger.Info("Connection pool health check started")
}

// DatabasePoolManager 数据库连接池管理器
type DatabasePoolManager struct {
	pools  map[string]Manager
	mu     sync.RWMutex
	logger logger.Logger
}

// NewDatabasePoolManager 创建数据库连接池管理器
func NewDatabasePoolManager(logger logger.Logger) *DatabasePoolManager {
	return &DatabasePoolManager{
		pools:  make(map[string]Manager),
		logger: logger,
	}
}

// CreatePool 创建连接池
func (m *DatabasePoolManager) CreatePool(name string, cfg *config.DatabaseConfig) (Manager, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[name]; exists {
		return nil, fmt.Errorf("pool %s already exists", name)
	}

	pool := NewConnectionPool(cfg, m.logger)
	m.pools[name] = pool

	// 启动健康检查
	if connPool, ok := pool.(*ConnectionPool); ok {
		connPool.StartHealthCheck()
	}

	m.logger.WithField("pool_name", name).Info("Database connection pool created")
	return pool, nil
}

// GetPool 获取连接池
func (m *DatabasePoolManager) GetPool(name string) (Manager, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, exists := m.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", name)
	}

	return pool, nil
}

// RemovePool 移除连接池
func (m *DatabasePoolManager) RemovePool(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pool, exists := m.pools[name]
	if !exists {
		return fmt.Errorf("pool %s not found", name)
	}

	if err := pool.Close(); err != nil {
		m.logger.WithError(err).WithField("pool_name", name).Error("Failed to close pool")
	}

	delete(m.pools, name)
	m.logger.WithField("pool_name", name).Info("Database connection pool removed")
	return nil
}

// GetAllStats 获取所有连接池的统计信息
func (m *DatabasePoolManager) GetAllStats() map[string]PoolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]PoolStats)
	for name, pool := range m.pools {
		stats[name] = pool.GetStats()
	}

	return stats
}

// HealthCheckAll 对所有连接池执行健康检查
func (m *DatabasePoolManager) HealthCheckAll(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]error)
	for name, pool := range m.pools {
		results[name] = pool.HealthCheck(ctx)
	}

	return results
}

// CloseAll 关闭所有连接池
func (m *DatabasePoolManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, pool := range m.pools {
		if err := pool.Close(); err != nil {
			m.logger.WithError(err).WithField("pool_name", name).Error("Failed to close pool")
			lastErr = err
		}
	}

	m.pools = make(map[string]Manager)
	m.logger.Info("All database connection pools closed")
	return lastErr
}
