// Package main ServerStatus data server
//
// @title ServerStatus API
// @version 1.0
// @description ServerStatus 监控 API 文档
// @BasePath /api
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	_ "github.com/kanshan/ServerStatus/data-server/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type SystemInfo struct {
	Hostname      string             `json:"hostname"`
	SessionID     string             `json:"session_id,omitempty"` // UUID session标识
	Timestamp     time.Time          `json:"timestamp"`
	CPU           CPUInfo            `json:"cpu"`
	Memory        MemInfo            `json:"memory"`
	Disk          DiskInfo           `json:"disk"`
	Network       NetInfo            `json:"network"`
	GPU           GPUInfo            `json:"gpu"`  // 保持兼容性，主GPU信息
	GPUs          []GPUInfo          `json:"gpus"` // 所有GPU信息
	OS            OSInfo             `json:"os"`
	Temperature   TempInfo           `json:"temperature"`
	ProjectKey    string             `json:"project_key,omitempty"`
	UserResources []UserResourceInfo `json:"user_resources,omitempty"` // 用户资源使用信息
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	CoreCount    int     `json:"core_count"`
	ModelName    string  `json:"model_name"`
}

type MemInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetInfo struct {
	BytesSent   uint64         `json:"bytes_sent"`   // 总发送字节数
	BytesRecv   uint64         `json:"bytes_recv"`   // 总接收字节数
	PacketsSent uint64         `json:"packets_sent"` // 总发送包数
	PacketsRecv uint64         `json:"packets_recv"` // 总接收包数
	SpeedSent   float64        `json:"speed_sent"`   // 发送速率 (KB/s)
	SpeedRecv   float64        `json:"speed_recv"`   // 接收速率 (KB/s)
	Interfaces  []NetInterface `json:"interfaces"`   // 网卡详细信息
}

type NetInterface struct {
	Name        string   `json:"name"`         // 网卡名称
	BytesSent   uint64   `json:"bytes_sent"`   // 发送字节数
	BytesRecv   uint64   `json:"bytes_recv"`   // 接收字节数
	PacketsSent uint64   `json:"packets_sent"` // 发送包数
	PacketsRecv uint64   `json:"packets_recv"` // 接收包数
	SpeedSent   float64  `json:"speed_sent"`   // 发送速率 (KB/s)
	SpeedRecv   float64  `json:"speed_recv"`   // 接收速率 (KB/s)
	IsUp        bool     `json:"is_up"`        // 网卡状态
	MTU         int      `json:"mtu"`          // MTU
	Addrs       []string `json:"addrs"`        // IP地址列表
}

type GPUInfo struct {
	Name         string  `json:"name"`
	MemoryTotal  uint64  `json:"memory_total"`
	MemoryUsed   uint64  `json:"memory_used"`
	UsagePercent float64 `json:"usage_percent"`
	Temperature  float64 `json:"temperature"`
}

type OSInfo struct {
	Platform     string `json:"platform"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"` // 规范字段
	Arch         string `json:"arch"`         // 兼容旧字段
	Uptime       uint64 `json:"uptime"`
}

type TempInfo struct {
	CPUTemp float64            `json:"cpu_temp"`
	GPUTemp float64            `json:"gpu_temp"`
	Sensors map[string]float64 `json:"sensors"`
	MaxTemp float64            `json:"max_temp"`
	AvgTemp float64            `json:"avg_temp"`
}

// UserResourceInfo 用户资源使用信息
type UserResourceInfo struct {
	Username      string        `json:"username"`
	UID           uint32        `json:"uid"`
	ProcessCount  int           `json:"process_count"`
	CPUPercent    float64       `json:"cpu_percent"`
	MemoryMB      uint64        `json:"memory_mb"`
	MemoryPercent float64       `json:"memory_percent"`
	TopProcesses  []ProcessInfo `json:"top_processes"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      uint64  `json:"memory_mb"`
	MemoryPercent float64 `json:"memory_percent"`
	Status        string  `json:"status"`
	Cmdline       string  `json:"cmdline,omitempty"`
}

type ServerData struct {
	mu             sync.RWMutex
	servers        map[string]*ServerInfo // key: sessionID, value: ServerInfo (内存缓存，用于快速访问)
	uuidStatsCache map[string]interface{} // UUID统计缓存
	uuidCacheTime  time.Time              // 缓存更新时间
	uuidCacheMutex sync.RWMutex           // 缓存读写锁
	database       DBStore                // 数据库实例（接口，便于替换存储）
}

type ServerInfo struct {
	Latest      *SystemInfo   `json:"latest"`
	History     []*SystemInfo `json:"history"`
	LastSeen    time.Time     `json:"last_seen"`
	OwnerUserID int64         `json:"owner_user_id,omitempty"`
}

type ServerStatus struct {
	Hostname         string             `json:"hostname"`
	SessionID        string             `json:"session_id,omitempty"` // UUID session标识
	LastSeen         time.Time          `json:"last_seen"`
	Status           string             `json:"status"`
	CPUPercent       float64            `json:"cpu_percent"`
	MemoryPercent    float64            `json:"memory_percent"`
	DiskPercent      float64            `json:"disk_percent"`
	OS               string             `json:"os"`
	CPUTemp          float64            `json:"cpu_temp"`
	GPUTemp          float64            `json:"gpu_temp"` // 保持兼容性，主GPU温度
	GPUs             []GPUInfo          `json:"gpus"`     // 所有GPU信息
	MaxTemp          float64            `json:"max_temp"`
	NetworkSpeedSent float64            `json:"network_speed_sent"`       // 网络发送速率 (KB/s)
	NetworkSpeedRecv float64            `json:"network_speed_recv"`       // 网络接收速率 (KB/s)
	NetworkBytesSent uint64             `json:"network_bytes_sent"`       // 总发送字节数
	NetworkBytesRecv uint64             `json:"network_bytes_recv"`       // 总接收字节数
	UserResources    []UserResourceInfo `json:"user_resources,omitempty"` // 用户资源使用信息
}

// VisitorEvent 访客埋点事件
type VisitorEvent struct {
	ProjectKey string    `json:"project_key"`
	Domain     string    `json:"domain,omitempty"`
	PageURL    string    `json:"page_url"`
	Referrer   string    `json:"referrer"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
	SessionID  string    `json:"session_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// DailyVisitorStat 每日访客统计
type DailyVisitorStat struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// PageVisitorStat 页面访问统计
type PageVisitorStat struct {
	Page  string `json:"page"`
	Count int    `json:"count"`
}

// VisitorStats 访客统计响应
type VisitorStats struct {
	ProjectKey  string             `json:"project_key"`
	From        time.Time          `json:"from"`
	To          time.Time          `json:"to"`
	TotalVisits int                `json:"total_visits"`
	UniqueIPs   int                `json:"unique_ips"`
	Daily       []DailyVisitorStat `json:"daily"`
	TopPages    []PageVisitorStat  `json:"top_pages,omitempty"`
}

// AuthUser 返回给前端的用户信息
type AuthUser struct {
	ID        int64       `json:"id"`
	Login     string      `json:"login"`
	Name      string      `json:"name"`
	AvatarURL string      `json:"avatar_url"`
	Email     string      `json:"email"`
	Config    interface{} `json:"config,omitempty"`
}

type ServerConfig struct {
	ProjectKey        string         `json:"project_key"`
	ServerKey         string         `json:"server_key"`
	Host              string         `json:"host"`
	Port              string         `json:"port"`
	RequireAuth       bool           `json:"require_auth"`
	DataLimit         int            `json:"data_limit"`                     // 数据保留条数限制
	DataInterval      int            `json:"data_interval"`                  // 数据上报间隔(秒)
	DatabasePath      string         `json:"database_path"`                  // 数据库文件路径
	DatabaseDriver    string         `json:"database_driver"`                // 数据库驱动（sqlite, future: postgres/mysql）
	DatabaseConns     []DBConnConfig `json:"database_connections,omitempty"` // 多数据库配置，默认使用第一个
	EnableCompression bool           `json:"enable_compression"`             // 启用gzip压缩
	CompressionLevel  int            `json:"compression_level"`              // 压缩级别(1-9)
	EnableWebSocket   bool           `json:"enable_websocket"`               // 启用WebSocket实时推送
	EnableCache       bool           `json:"enable_cache"`                   // 启用Redis缓存
	RedisAddr         string         `json:"redis_addr"`                     // Redis地址
	RedisPassword     string         `json:"redis_password"`                 // Redis密码
	RedisDB           int            `json:"redis_db"`                       // Redis数据库编号
}

// DBConnConfig 单个数据库连接配置
type DBConnConfig struct {
	Driver string `json:"driver"`
	Path   string `json:"path"`
}

// validateServerConfig 验证服务器配置
func validateServerConfig(config *ServerConfig) error {
	var errors []string

	// 验证主机地址
	if config.Host != "" {
		if !isValidHost(config.Host) {
			errors = append(errors, fmt.Sprintf("无效的主机地址: %s", config.Host))
		}
	}

	// 验证端口
	if config.Port != "" {
		if !isValidPort(config.Port) {
			errors = append(errors, fmt.Sprintf("无效的端口号: %s", config.Port))
		}
	}

	// 验证数据限制
	if config.DataLimit < 0 {
		errors = append(errors, "数据限制不能为负数")
	}
	if config.DataLimit > 100000 {
		errors = append(errors, "数据限制不能超过100000条")
	}

	// 验证数据间隔
	if config.DataInterval <= 0 {
		errors = append(errors, "数据间隔必须大于0秒")
	}
	if config.DataInterval > 3600 {
		errors = append(errors, "数据间隔不能超过3600秒")
	}

	// 验证压缩级别
	if config.EnableCompression {
		if config.CompressionLevel < 1 || config.CompressionLevel > 9 {
			errors = append(errors, "压缩级别必须在1-9之间")
		}
	}

	// 验证Redis配置
	if config.EnableCache {
		if config.RedisAddr == "" {
			errors = append(errors, "启用缓存时必须指定Redis地址")
		} else if !isValidRedisAddr(config.RedisAddr) {
			errors = append(errors, fmt.Sprintf("无效的Redis地址: %s", config.RedisAddr))
		}

		if config.RedisDB < 0 || config.RedisDB > 15 {
			errors = append(errors, "Redis数据库编号必须在0-15之间")
		}
	}

	// 验证数据库路径
	if config.DatabasePath != "" {
		if !isValidDBPath(config.DatabaseDriver, config.DatabasePath) {
			errors = append(errors, fmt.Sprintf("无效的数据库路径/DSN: %s", config.DatabasePath))
		}
	}

	// 验证数据库驱动（允许预置 pg/mysql 以便 docker 测试配置，但当前实现仅 sqlite）
	if config.DatabaseDriver == "" {
		config.DatabaseDriver = "sqlite"
	}
	if !isSupportedDriver(config.DatabaseDriver) {
		errors = append(errors, fmt.Sprintf("不支持的数据库驱动: %s (当前实现仅 sqlite，pg/mysql 需后续实现)", config.DatabaseDriver))
	}

	// 验证多数据库列表（若存在，默认取第一个）
	for i, c := range config.DatabaseConns {
		driver := strings.ToLower(c.Driver)
		if driver == "" {
			driver = strings.ToLower(config.DatabaseDriver)
		}
		if c.Path != "" && !isValidDBPath(driver, c.Path) {
			errors = append(errors, fmt.Sprintf("无效的数据库路径/DSN(database_connections[%d]): %s", i, c.Path))
		}
		if !isSupportedDriver(driver) {
			errors = append(errors, fmt.Sprintf("不支持的数据库驱动(database_connections[%d]): %s (当前实现仅 sqlite)", i, c.Driver))
		}
	}

	// 验证密钥格式
	if config.ProjectKey != "" && config.ProjectKey != "public" && len(config.ProjectKey) < 8 {
		errors = append(errors, "项目密钥长度不能少于8个字符")
	}

	if config.ServerKey != "" && config.ServerKey != "serverstatus.ltd" && len(config.ServerKey) < 8 {
		errors = append(errors, "服务器密钥长度不能少于8个字符")
	}

	if len(errors) > 0 {
		return fmt.Errorf("配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// isValidHost 验证主机地址格式
func isValidHost(host string) bool {
	// 允许IP地址或域名
	// IPv4地址
	if net.ParseIP(host) != nil {
		return true
	}

	// 简单的域名验证
	if host == "localhost" || host == "0.0.0.0" {
		return true
	}

	// 域名格式检查
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`, host); matched {
		return true
	}

	return false
}

// isValidPort 验证端口号格式
func isValidPort(port string) bool {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return portNum >= 1 && portNum <= 65535
}

// isValidRedisAddr 验证Redis地址格式
func isValidRedisAddr(addr string) bool {
	// 支持 host:port 格式
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}

	host := parts[0]
	port := parts[1]

	if !isValidHost(host) {
		return false
	}

	return isValidPort(port)
}

// isValidPath 验证路径格式
func isValidPath(path string) bool {
	// 检查是否包含非法字符
	if strings.ContainsAny(path, "<>:\"|?*") {
		return false
	}

	// 检查路径长度
	if len(path) > 260 {
		return false
	}

	return true
}

// isValidDBPath 依据驱动校验路径/DSN
func isValidDBPath(driver, path string) bool {
	if path == "" {
		return true
	}
	drv := strings.ToLower(driver)
	// sqlite 仍按路径规则校验
	if drv == "" || drv == "sqlite" || drv == "sqlite3" {
		return isValidPath(path)
	}
	// 允许常见 DSN scheme
	if strings.HasPrefix(path, "postgres://") || strings.HasPrefix(path, "postgresql://") ||
		strings.HasPrefix(path, "pg://") || strings.HasPrefix(path, "mysql://") || strings.HasPrefix(path, "mariadb://") {
		return true
	}
	// 兜底：接受其它非空值
	return isValidPath(path)
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getClientIP 获取真实客户端IP
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// getDomainFromURL 从URL提取域名
func getDomainFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	if host == "" {
		return ""
	}
	return strings.ToLower(host)
}

// newOAuthState 生成并存储state
func newOAuthState() string {
	state := generateRandomHex(16)
	githubStateStore.mu.Lock()
	githubStateStore.states[state] = time.Now().Add(5 * time.Minute)
	githubStateStore.mu.Unlock()
	return state
}

// validateOAuthState 校验state有效性
func validateOAuthState(state string) bool {
	githubStateStore.mu.Lock()
	defer githubStateStore.mu.Unlock()
	exp, ok := githubStateStore.states[state]
	if !ok {
		return false
	}
	delete(githubStateStore.states, state)
	return time.Now().Before(exp)
}

// generateRandomHex 生成随机HEX字符串
func generateRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// setSessionCookie 设置会话cookie
func setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})
}

// clearSessionCookie 清除会话cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
}

// getSessionUser 获取当前会话用户（不写cookie）
func getSessionUser(r *http.Request) (*User, error) {
	if data.database == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, nil
	}
	session, err := data.database.GetSession(cookie.Value)
	if err != nil || session == nil {
		return nil, nil
	}
	if time.Now().After(session.ExpiresAt) {
		_ = data.database.DeleteSession(cookie.Value)
		return nil, nil
	}
	user, err := data.database.GetUserByID(session.UserID)
	if err != nil || user == nil {
		return nil, nil
	}
	return user, nil
}

// isPublicPath 判断是否无需登录
func isPublicPath(path string) bool {
	if publicPaths[path] {
		return true
	}
	for _, p := range publicPathPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// authMiddleware 除上报、登录等公共路径外，强制登录
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 若未开启认证，直接放行
		if !serverConfig.RequireAuth {
			next.ServeHTTP(w, r)
			return
		}

		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		user, _ := getSessionUser(r)
		if user == nil {
			http.Error(w, "需要登录", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AccessKey缓存结构
type AccessKeyCache struct {
	mu    sync.RWMutex
	cache map[string]string // key: serverKey:projectKey, value: accessKey
}

var (
	startTime = time.Now() // 服务器启动时间

	buildVersion = "2.2.0-dev" // 编译时可通过 -ldflags 覆盖
	buildCommit  = "unknown"
	buildTime    = "unknown"

	// 1x1 透明GIF，用于访客埋点像素返回
	trackingPixel = []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0, 255, 255, 255, 33, 249, 4, 1, 0, 0, 1, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	// GitHub OAuth 配置（从环境变量读取）
	githubClientID     = os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	githubCallbackURL  = os.Getenv("GITHUB_CALLBACK_URL")

	// 会话
	sessionCookieName = "ss_session"
	sessionTTL        = 7 * 24 * time.Hour

	// GitHub state 防重放
	githubStateStore = struct {
		mu     sync.Mutex
		states map[string]time.Time
	}{states: make(map[string]time.Time)}

	// 公共路径（无需登录）
	publicPaths = map[string]bool{
		"/api/data":                 true,
		"/api/auth/github/login":    true,
		"/api/auth/github/callback": true,
		"/api/health":               true,
		"/api/version":              true,
		"/api/stats":                true,
	}

	publicPathPrefixes = []string{
		"/download/",
		"/install",
		"/API.md",
	}

	data = &ServerData{
		servers:        make(map[string]*ServerInfo),
		uuidStatsCache: make(map[string]interface{}),
		uuidCacheTime:  time.Time{}, // 零值表示未初始化
		database:       nil,         // 将在main函数中初始化
	}

	serverConfig = ServerConfig{
		ProjectKey:        "public",           // 默认项目密钥
		ServerKey:         "serverstatus.ltd", // 默认服务器密钥
		Host:              "0.0.0.0",
		Port:              "8080",
		RequireAuth:       true,
		DataLimit:         1000,                     // 默认保留1000条数据
		DataInterval:      5,                        // 默认5秒间隔
		DatabasePath:      "./data/serverstatus.db", // 数据库路径
		DatabaseDriver:    "sqlite",                 // 默认驱动
		DatabaseConns:     []DBConnConfig{},         // 可选多DB配置
		EnableCompression: true,                     // 默认启用压缩
		CompressionLevel:  6,                        // 默认压缩级别
		EnableWebSocket:   true,                     // 默认启用WebSocket
		EnableCache:       true,                     // 默认启用缓存
		RedisAddr:         "localhost:6379",         // 默认Redis地址
		RedisPassword:     "",                       // 默认无密码
		RedisDB:           0,                        // 默认数据库
	}

	// 全局AccessKey缓存
	accessKeyCache = &AccessKeyCache{
		cache: make(map[string]string),
	}

	// 命令行参数
	projectKey        = flag.String("key", "", "项目认证密钥")
	serverKey         = flag.String("server-key", "", "服务器密钥 (用于双密钥认证)")
	host              = flag.String("host", "0.0.0.0", "服务器绑定IP地址")
	port              = flag.String("port", "8080", "服务器端口")
	configFile        = flag.String("config", "server-config.json", "服务器配置文件路径")
	requireAuth       = flag.Bool("auth", false, "是否要求API密钥认证")
	dataLimit         = flag.Int("data-limit", 1000, "数据保留条数限制")
	dataInterval      = flag.Int("data-interval", 5, "推荐的数据上报间隔(秒)")
	enableCompression = flag.Bool("compression", true, "启用gzip压缩")
	compressionLevel  = flag.Int("compression-level", 6, "gzip压缩级别(1-9)")
	enableWebSocket   = flag.Bool("websocket", true, "启用WebSocket实时推送")
	enableCache       = flag.Bool("cache", true, "启用Redis缓存")
	redisAddr         = flag.String("redis", "localhost:6379", "Redis服务器地址")
	redisPassword     = flag.String("redis-password", "", "Redis服务器密码")
	redisDB           = flag.Int("redis-db", 0, "Redis数据库编号")
	databaseDriver    = flag.String("db-driver", "sqlite", "数据库驱动 (sqlite, 未来: postgres/mysql)")
	showHelp          = flag.Bool("help", false, "显示帮助信息")
)

// 移除嵌入的前端文件系统，实现前后端分离

const (
	offlineThreshold = 30 * time.Second
)

// corsMiddleware 处理CORS跨域请求
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Server-Key, X-Project-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24小时

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

// compressionMiddleware 为API响应添加gzip压缩
func compressionMiddleware(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(next)
}

// compressionConditionalMiddleware 为指定路径选择性添加压缩
func compressionConditionalMiddleware(paths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查请求路径是否需要压缩
			shouldCompress := false
			for _, path := range paths {
				if strings.HasPrefix(r.URL.Path, path) {
					shouldCompress = true
					break
				}
			}

			if shouldCompress {
				// 检查客户端是否支持gzip
				if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
					gziphandler.GzipHandler(next).ServeHTTP(w, r)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	flag.Parse()

	if *showHelp {
		printServerUsage()
		return
	}

	// 加载配置文件
	loadServerConfig()

	// 命令行参数覆盖配置文件
	if *projectKey != "" {
		serverConfig.ProjectKey = *projectKey
	}
	if *serverKey != "" {
		serverConfig.ServerKey = *serverKey
	}
	if *host != "0.0.0.0" {
		serverConfig.Host = *host
	}
	if *port != "8080" {
		serverConfig.Port = *port
	}
	if *requireAuth {
		serverConfig.RequireAuth = true
	}
	if *dataLimit != 1000 {
		serverConfig.DataLimit = *dataLimit
	}
	if *dataInterval != 5 {
		serverConfig.DataInterval = *dataInterval
	}
	if *enableCompression != serverConfig.EnableCompression {
		serverConfig.EnableCompression = *enableCompression
	}
	if *compressionLevel >= 1 && *compressionLevel <= 9 {
		serverConfig.CompressionLevel = *compressionLevel
	}
	if *enableWebSocket != serverConfig.EnableWebSocket {
		serverConfig.EnableWebSocket = *enableWebSocket
	}
	if *enableCache != serverConfig.EnableCache {
		serverConfig.EnableCache = *enableCache
	}
	if *redisAddr != serverConfig.RedisAddr {
		serverConfig.RedisAddr = *redisAddr
	}
	if *redisPassword != serverConfig.RedisPassword {
		serverConfig.RedisPassword = *redisPassword
	}
	if *redisDB != serverConfig.RedisDB {
		serverConfig.RedisDB = *redisDB
	}
	if *databaseDriver != "" && strings.ToLower(*databaseDriver) != strings.ToLower(serverConfig.DatabaseDriver) {
		serverConfig.DatabaseDriver = strings.ToLower(*databaseDriver)
	}

	// 初始化缓存管理器
	if serverConfig.EnableCache {
		log.Println("初始化缓存管理器...")
		cacheManager = NewCacheManager(serverConfig.RedisAddr, serverConfig.RedisPassword, serverConfig.RedisDB)

		// 启动缓存清理协程
		go func() {
			ticker := time.NewTicker(10 * time.Minute) // 每10分钟清理一次过期缓存
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					cacheManager.cleanupExpiredCache()
				}
			}
		}()
	} else {
		cacheManager = NewCacheManager("", "", 0) // 禁用缓存
	}

	// 初始化数据库
	log.Println("初始化数据库...")
	db, err := initDBStore()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	data.database = db

	// 加载现有服务器数据到内存缓存
	if err := loadServersFromDatabase(); err != nil {
		log.Printf("从数据库加载服务器数据失败: %v", err)
	}

	log.Println("初始化健康检查服务...")
	initializeHealthService(db)

	log.Println("启动 ServerStatus Monitor Data Server...")
	log.Printf("端口: %s", serverConfig.Port)
	log.Printf("数据限制: %d 条记录", serverConfig.DataLimit)
	log.Printf("推荐数据间隔: %d 秒", serverConfig.DataInterval)
	activeDriver, activePath := getActiveDBConfig()
	log.Printf("数据库驱动: %s", activeDriver)
	log.Printf("数据库路径: %s", activePath)
	if serverConfig.EnableCompression {
		log.Printf("gzip压缩: 启用 (级别: %d)", serverConfig.CompressionLevel)
	} else {
		log.Println("gzip压缩: 禁用")
	}
	if serverConfig.EnableWebSocket {
		log.Println("WebSocket实时推送: 启用")
	} else {
		log.Println("WebSocket实时推送: 禁用")
	}
	if serverConfig.EnableCache {
		log.Printf("Redis缓存: 启用 (地址: %s)", serverConfig.RedisAddr)
	} else {
		log.Println("Redis缓存: 禁用")
	}
	if serverConfig.RequireAuth {
		log.Println("API认证: 启用 (双密钥认证模式)")
	} else {
		log.Println("API认证: 禁用")
	}

	// GitHub 回调默认值
	if githubCallbackURL == "" {
		host := serverConfig.Host
		if host == "0.0.0.0" {
			host = "localhost"
		}
		githubCallbackURL = fmt.Sprintf("http://%s:%s/api/auth/github/callback", host, serverConfig.Port)
	}

	r := mux.NewRouter()

	// 添加CORS中间件支持前后端分离
	r.Use(corsMiddleware)

	// Swagger 文档
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// 定义需要压缩的API路径
	compressiblePaths := []string{
		"/api/servers",
		"/api/server/",
		"/api/access/",
		"/api/uuid-count",
		"/api/user-resources/",
		"/api/visitor/stats",
		"/api/visitor/aggregate",
		"/api/auth/me",
	}

	// API路由 - 应用压缩
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)
	api.Use(compressionConditionalMiddleware(compressiblePaths))

	api.HandleFunc("/data", handleData).Methods("POST")
	api.HandleFunc("/register-session", handleRegisterSession).Methods("POST")
	api.HandleFunc("/servers", handleGetServers).Methods("GET")
	api.HandleFunc("/server/{sessionID}", handleGetServer).Methods("GET")
	// 双密钥认证相关路由
	api.HandleFunc("/generate-access-key", handleGenerateAccessKey).Methods("POST")
	api.HandleFunc("/access/{accessKey}/servers", handleGetServersByAccessKey).Methods("GET")
	api.HandleFunc("/access/{accessKey}/server-by-session/{sessionID}", handleGetServerBySessionID).Methods("GET")
	api.HandleFunc("/access/{accessKey}/user-resources/{sessionID}", handleGetUserResourcesByAccessKey).Methods("GET")
	api.HandleFunc("/user-resources/{sessionID}", handleGetUserResources).Methods("GET")
	api.HandleFunc("/uuid-count", handleGetUUIDCount).Methods("GET")
	api.HandleFunc("/visitor/track", handleTrackVisit).Methods("GET")
	api.HandleFunc("/visitor/stats", handleVisitorStats).Methods("GET")
	api.HandleFunc("/visitor/aggregate", handleVisitorAggregate).Methods("GET")
	api.HandleFunc("/visitor/bindings", handleListDomainBindings).Methods("GET")
	api.HandleFunc("/visitor/bindings", handleUpsertDomainBinding).Methods("POST")

	// 健康检查和系统状态路由
	api.HandleFunc("/health", handleHealth).Methods("GET")
	api.HandleFunc("/version", handleVersion).Methods("GET")
	api.HandleFunc("/stats", handleSystemStats).Methods("GET")

	// 配置管理路由
	api.HandleFunc("/reload-config", handleReloadConfig).Methods("POST")

	// 数据导出路由
	api.HandleFunc("/export/servers", handleExportServersCSV).Methods("GET")
	api.HandleFunc("/export/history", handleExportHistoryCSV).Methods("GET")
	api.HandleFunc("/export/user-resources", handleExportUserResourcesCSV).Methods("GET")
	// 部署辅助：返回一键运行客户端的命令与env
	api.HandleFunc("/deploy/agent-command", handleDeployAgentCommand).Methods("GET")
	// GitHub OAuth
	api.HandleFunc("/auth/github/login", handleAuthGitHubLogin).Methods("GET")
	api.HandleFunc("/auth/github/callback", handleAuthGitHubCallback).Methods("GET")
	api.HandleFunc("/auth/me", handleAuthMe).Methods("GET")
	api.HandleFunc("/auth/logout", handleAuthLogout).Methods("POST")

	// WebSocket实时通信路由
	if serverConfig.EnableWebSocket {
		r.HandleFunc("/ws", webSocketManager.handleWebSocket).Methods("GET")
		r.HandleFunc("/api/ws-stats", handleWebSocketStats).Methods("GET")
	}

	// 缓存统计路由
	if serverConfig.EnableCache {
		r.HandleFunc("/api/cache-stats", handleCacheStats).Methods("GET")
	}

	// 下载路由（不压缩）
	r.HandleFunc("/download/{filename}", handleDownload).Methods("GET")
	r.HandleFunc("/install", handleInstallScript).Methods("GET")

	// API文档服务（不压缩）
	r.HandleFunc("/API.md", handleAPIDoc).Methods("GET")

	// 前后端分离：移除静态文件服务
	// 前端需要独立部署，通过API接口访问数据

	// 启动WebSocket管理器
	if serverConfig.EnableWebSocket {
		webSocketManager.Start()
		log.Println("WebSocket管理器已启动")
	}

	// 启动清理协程
	go cleanupRoutine()

	log.Printf("API服务器启动在 %s:%s", serverConfig.Host, serverConfig.Port)
	if serverConfig.Host == "0.0.0.0" {
		log.Printf("API基础地址: http://localhost:%s/api", serverConfig.Port)
		log.Printf("API文档地址: http://localhost:%s/API.md", serverConfig.Port)
		if serverConfig.EnableWebSocket {
			log.Printf("WebSocket地址: ws://localhost:%s/ws", serverConfig.Port)
		}
	} else {
		log.Printf("API基础地址: http://%s:%s/api", serverConfig.Host, serverConfig.Port)
		log.Printf("API文档地址: http://%s:%s/API.md", serverConfig.Host, serverConfig.Port)
		if serverConfig.EnableWebSocket {
			log.Printf("WebSocket地址: ws://%s:%s/ws", serverConfig.Host, serverConfig.Port)
		}
	}
	log.Println("前后端已分离，前端需独立部署")
	log.Fatal(http.ListenAndServe(serverConfig.Host+":"+serverConfig.Port, r))
}

// loadServersFromDatabase 从数据库加载服务器数据到内存缓存
func loadServersFromDatabase() error {
	if data.database == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 预估数量后一次拉取，避免多次分页 IO
	total, err := data.database.GetServerCount("")
	if err != nil {
		return err
	}
	if total == 0 {
		log.Println("数据库无服务器记录，无需预加载")
		return nil
	}

	servers, err := data.database.GetAllServers("", 0, total)
	if err != nil {
		return err
	}

	loaded := 0
	data.mu.Lock()
	defer data.mu.Unlock()

	for _, srv := range servers {
		if srv == nil || srv.Latest == nil || srv.Latest.SessionID == "" {
			continue
		}

		// 加载历史数据（按 DataLimit 限制）
		history, err := data.database.GetHistoryData(srv.Latest.SessionID, serverConfig.DataLimit)
		if err != nil {
			log.Printf("预加载历史数据失败: %v", err)
			history = nil
		} else {
			reverseHistory(history)
		}

		data.servers[srv.Latest.SessionID] = &ServerInfo{
			Latest:   srv.Latest,
			History:  history,
			LastSeen: srv.LastSeen,
		}
		loaded++
	}

	log.Printf("从数据库预加载服务器数据完成: %d/%d 台", loaded, total)
	return nil
}

// initializeHealthService 初始化健康检查服务
func initializeHealthService(db DBStore) {
	log.Println("✅ 健康检查服务已初始化 (使用简化实现)")
}

// saveServerToDatabase 保存服务器数据到数据库
func saveServerToDatabase(sessionID, hostname, projectKey string, info *SystemInfo) error {
	if data.database == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 保存服务器信息
	if err := data.database.SaveServerInfo(sessionID, hostname, projectKey, info); err != nil {
		return err
	}

	// 保存历史数据
	if err := data.database.SaveHistoryData(sessionID, hostname, projectKey, info); err != nil {
		return err
	}

	return nil
}

// handleData godoc
// @Summary 上报监控数据
// @Description 监控代理上报系统信息，支持双密钥校验
// @Tags data
// @Accept json
// @Produce json
// @Param X-Project-Key header string false "项目密钥（默认 public）"
// @Param X-Server-Key header string false "服务器密钥（RequireAuth=true 时需要）"
// @Param payload body SystemInfo true "系统信息"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "解析数据失败"
// @Failure 401 {string} string "无效的服务器密钥"
// @Failure 500 {string} string "数据保存失败"
// @Router /data [post]
func handleData(w http.ResponseWriter, r *http.Request) {
	log.Printf("[数据上报] 收到数据上报请求，来源IP: %s", r.RemoteAddr)

	var projectKey string
	// 服务器密钥验证
	if serverConfig.RequireAuth {
		serverKey := r.Header.Get("X-Server-Key")
		log.Printf("[数据上报] 验证配置 - RequireAuth: %v, ServerKey: %s, 收到ServerKey: %s", serverConfig.RequireAuth, serverConfig.ServerKey, serverKey)

		// 只验证ServerKey
		if serverConfig.ServerKey != "" && serverKey != serverConfig.ServerKey {
			log.Printf("[数据上报] 验证失败: 服务器密钥不匹配 - 收到: %s, 期望: %s", serverKey, serverConfig.ServerKey)
			http.Error(w, "无效的服务器密钥", http.StatusUnauthorized)
			return
		}
		log.Printf("[数据上报] 服务器密钥验证通过")
	}

	// 获取ProjectKey用于数据分组（用户自定义，不验证）
	projectKey = r.Header.Get("X-Project-Key")
	if projectKey == "" {
		projectKey = "default" // 默认密钥组
	}
	log.Printf("[数据上报] ProjectKey: %s", projectKey)

	var info SystemInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, "解析数据失败", http.StatusBadRequest)
		return
	}

	// 为数据添加项目密钥标识
	info.ProjectKey = projectKey

	// 统一确定session key，优先使用客户端提供的sessionID，否则根据指纹生成
	originalSession := info.SessionID
	sessionKey := deriveSessionKey(&info, projectKey, r.RemoteAddr)
	if originalSession == "" && sessionKey != "" {
		log.Printf("[数据上报] 未提供SessionID，为主机 %s 生成自动Session: %s", info.Hostname, sessionKey)
	}
	info.SessionID = sessionKey

	// 保存到数据库
	if err := saveServerToDatabase(sessionKey, info.Hostname, projectKey, &info); err != nil {
		log.Printf("保存数据到数据库失败: %v", err)
		http.Error(w, "数据保存失败", http.StatusInternalServerError)
		return
	}

	data.mu.Lock()
	defer data.mu.Unlock()

	if data.servers[sessionKey] == nil {
		data.servers[sessionKey] = &ServerInfo{
			History: make([]*SystemInfo, 0, serverConfig.DataLimit),
		}
		log.Printf("新服务器注册: %s (Session: %s)", info.Hostname, sessionKey)
	}

	server := data.servers[sessionKey]
	server.Latest = &info
	server.LastSeen = time.Now()

	// 添加到内存历史记录（保持向后兼容）
	server.History = append(server.History, &info)
	if len(server.History) > serverConfig.DataLimit {
		server.History = server.History[1:]
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("收到 %s 的数据上报 (Session: %s)", info.Hostname, sessionKey)

	// WebSocket实时推送服务器更新
	if serverConfig.EnableWebSocket {
		// 检查是否是新服务器或状态变化
		isNewServer := data.servers[sessionKey] == nil || len(data.servers[sessionKey].History) <= 1
		action := "update"
		if isNewServer {
			action = "online"
		}

		// 构造服务器状态
		serverStatus := ServerStatus{
			Hostname:         info.Hostname,
			SessionID:        info.SessionID,
			LastSeen:         time.Now(),
			Status:           "online",
			CPUPercent:       info.CPU.UsagePercent,
			MemoryPercent:    info.Memory.UsagePercent,
			DiskPercent:      info.Disk.UsagePercent,
			OS:               info.OS.Platform,
			CPUTemp:          info.Temperature.CPUTemp,
			GPUTemp:          info.Temperature.GPUTemp,
			GPUs:             info.GPUs,
			MaxTemp:          info.Temperature.MaxTemp,
			NetworkSpeedSent: info.Network.SpeedSent,
			NetworkSpeedRecv: info.Network.SpeedRecv,
			NetworkBytesSent: info.Network.BytesSent,
			NetworkBytesRecv: info.Network.BytesRecv,
			UserResources:    info.UserResources,
		}

		// 广播更新到所有WebSocket客户端
		webSocketManager.BroadcastServerUpdate(serverStatus, action)
	}
}

// handleGetServers godoc
// @Summary 获取服务器列表
// @Tags server
// @Produce json
// @Success 200 {array} ServerStatus
// @Router /servers [get]
func handleGetServers(w http.ResponseWriter, r *http.Request) {
	user, _ := getSessionUser(r)
	if user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}
	// 解析分页参数
	page := 1
	limit := 100
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	var servers []ServerStatus
	now := time.Now()

	// 从缓存获取项目密钥
	projectKey := r.URL.Query().Get("project_key")
	if projectKey == "" {
		projectKey = "public"
	}

	// 尝试从缓存获取数据（仅在第一页且无特殊查询时使用缓存）
	ctx := r.Context()
	useCache := page == 1 && limit <= 100 && r.URL.Query().Get("search") == ""

	var totalCount int
	var fromCache bool

	if useCache && serverConfig.EnableCache {
		// 尝试从缓存获取服务器列表
		cachedServers, err := cacheManager.GetServersList(ctx, projectKey)
		if err == nil && len(cachedServers) > 0 {
			servers = cachedServers
			fromCache = true

			// 尝试从缓存获取总数
			cachedCount, err := cacheManager.GetServerCount(ctx, projectKey)
			if err == nil {
				totalCount = cachedCount
			}

			log.Printf("从缓存获取服务器列表: %d 个服务器", len(servers))
		}
	}

	// 如果缓存未命中，从数据库获取数据
	if !fromCache {
		if data.database != nil {
			// 从数据库获取服务器数据
			dbServers, err := data.database.GetAllServers(projectKey, offset, limit)
			if err != nil {
				log.Printf("从数据库获取服务器列表失败: %v", err)
				// 降级到内存数据
				servers = getServersFromMemory(projectKey, now, user)
			} else {
				// 转换数据库数据为ServerStatus格式
				for _, server := range dbServers {
					if server.Latest == nil {
						continue
					}

					status := "online"
					if now.Sub(server.Latest.Timestamp) > offlineThreshold {
						status = "offline"
					}

					servers = append(servers, ServerStatus{
						Hostname:         server.Latest.Hostname,
						SessionID:        server.Latest.SessionID,
						LastSeen:         server.LastSeen,
						Status:           status,
						CPUPercent:       server.Latest.CPU.UsagePercent,
						MemoryPercent:    server.Latest.Memory.UsagePercent,
						DiskPercent:      server.Latest.Disk.UsagePercent,
						OS:               server.Latest.OS.Platform,
						CPUTemp:          server.Latest.Temperature.CPUTemp,
						GPUTemp:          server.Latest.Temperature.GPUTemp,
						GPUs:             server.Latest.GPUs,
						MaxTemp:          server.Latest.Temperature.MaxTemp,
						NetworkSpeedSent: server.Latest.Network.SpeedSent,
						NetworkSpeedRecv: server.Latest.Network.SpeedRecv,
						NetworkBytesSent: server.Latest.Network.BytesSent,
						NetworkBytesRecv: server.Latest.Network.BytesRecv,
						UserResources:    server.Latest.UserResources,
					})
				}

				// 获取总数用于分页
				totalCount, _ = data.database.GetServerCount(projectKey)

				// 缓存查询结果（仅缓存第一页）
				if useCache && serverConfig.EnableCache && page == 1 {
					go func() {
						cacheCtx := context.Background()
						_ = cacheManager.SetServersList(cacheCtx, projectKey, servers)
						_ = cacheManager.SetServerCount(cacheCtx, projectKey, totalCount)
						log.Printf("服务器列表已缓存: %d 个服务器", len(servers))
					}()
				}
			}
		} else {
			// 从内存获取数据
			servers = getServersFromMemory(projectKey, now, user)
		}
	}

	// 按主机名排序
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})

	// 设置分页头
	w.Header().Set("X-Total-Count", fmt.Sprintf("%d", totalCount))
	w.Header().Set("X-Page", fmt.Sprintf("%d", page))
	w.Header().Set("X-Per-Page", fmt.Sprintf("%d", limit))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(servers); err != nil {
		log.Printf("Error encoding servers response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// getServersFromMemory 从内存获取服务器数据（降级方案）
func getServersFromMemory(projectKey string, now time.Time, user *User) []ServerStatus {
	data.mu.RLock()
	defer data.mu.RUnlock()

	var userID int64
	if user != nil {
		userID = user.ID
	}

	var servers []ServerStatus

	for _, server := range data.servers {
		if server.Latest == nil {
			continue
		}

		if server.OwnerUserID > 0 && server.OwnerUserID != userID {
			continue
		}

		// 根据项目密钥过滤
		if projectKey != "public" && server.Latest.ProjectKey != projectKey {
			continue
		}

		status := "online"
		if now.Sub(server.Latest.Timestamp) > offlineThreshold {
			status = "offline"
		}

		servers = append(servers, ServerStatus{
			Hostname:         server.Latest.Hostname,
			SessionID:        server.Latest.SessionID,
			LastSeen:         server.LastSeen,
			Status:           status,
			CPUPercent:       server.Latest.CPU.UsagePercent,
			MemoryPercent:    server.Latest.Memory.UsagePercent,
			DiskPercent:      server.Latest.Disk.UsagePercent,
			OS:               server.Latest.OS.Platform,
			CPUTemp:          server.Latest.Temperature.CPUTemp,
			GPUTemp:          server.Latest.Temperature.GPUTemp,
			GPUs:             server.Latest.GPUs,
			MaxTemp:          server.Latest.Temperature.MaxTemp,
			NetworkSpeedSent: server.Latest.Network.SpeedSent,
			NetworkSpeedRecv: server.Latest.Network.SpeedRecv,
			NetworkBytesSent: server.Latest.Network.BytesSent,
			NetworkBytesRecv: server.Latest.Network.BytesRecv,
			UserResources:    server.Latest.UserResources,
		})
	}

	return servers
}

// handleGetServer godoc
// @Summary 获取指定服务器详情
// @Tags server
// @Produce json
// @Param sessionID path string true "SessionID"
// @Success 200 {object} ServerInfo
// @Failure 404 {string} string "服务器不存在"
// @Router /server/{sessionID} [get]
func handleGetServer(w http.ResponseWriter, r *http.Request) {
	user, _ := getSessionUser(r)
	if user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["sessionID"]

	server := getServerBySessionID(sessionID)
	if server == nil {
		http.Error(w, "服务器不存在", http.StatusNotFound)
		return
	}

	if server.OwnerUserID > 0 && server.OwnerUserID != user.ID {
		http.Error(w, "无权访问该服务器", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(server); err != nil {
		log.Printf("Error encoding server response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// 已移除基于项目密钥和访问令牌的处理函数，只保留AccessKey访问方式

// AccessKeyRequest 生成访问密钥请求结构
type AccessKeyRequest struct {
	ServerKey  string `json:"server_key"`
	ProjectKey string `json:"project_key"`
}

// AccessKeyResponse 生成访问密钥响应结构
type AccessKeyResponse struct {
	AccessKey  string `json:"access_key"`
	ProjectKey string `json:"project_key"`
	Message    string `json:"message"`
}

// SessionRegisterRequest session注册请求结构
type SessionRegisterRequest struct {
	Hostname   string `json:"hostname"`
	ProjectKey string `json:"project_key"`
}

// SessionRegisterResponse session注册响应结构
type SessionRegisterResponse struct {
	SessionID string `json:"session_id"`
	Hostname  string `json:"hostname"`
}

// handleRegisterSession godoc
// @Summary 注册新的 Session
// @Tags session
// @Accept json
// @Produce json
// @Param payload body SessionRegisterRequest true "主机与项目"
// @Success 200 {object} SessionRegisterResponse
// @Failure 400 {string} string "参数错误"
// @Failure 401 {string} string "项目密钥验证失败"
// @Router /register-session [post]
func handleRegisterSession(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Session注册] 收到注册请求，来源IP: %s", r.RemoteAddr)

	user, _ := getSessionUser(r)
	if user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	var req SessionRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Session注册] JSON解析失败: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[Session注册] 请求内容 - Hostname: %s, ProjectKey: %s", req.Hostname, req.ProjectKey)

	if req.Hostname == "" {
		log.Printf("[Session注册] 验证失败: Hostname为空")
		http.Error(w, "Hostname is required", http.StatusBadRequest)
		return
	}

	if req.ProjectKey == "" {
		log.Printf("[Session注册] 验证失败: ProjectKey为空")
		http.Error(w, "Project key is required", http.StatusBadRequest)
		return
	}

	// 验证项目密钥
	log.Printf("[Session注册] 验证配置 - RequireAuth: %v, ServerProjectKey: %s", serverConfig.RequireAuth, serverConfig.ProjectKey)
	if serverConfig.RequireAuth && !isValidProjectKey(req.ProjectKey) {
		log.Printf("[Session注册] 验证失败: 项目密钥验证不通过 - 收到: %s, 期望: %s", req.ProjectKey, serverConfig.ProjectKey)
		http.Error(w, "Invalid project key", http.StatusUnauthorized)
		return
	}

	log.Printf("[Session注册] 验证成功，生成Session ID")
	// 生成UUID作为session ID
	sessionID := generateUUID()

	data.mu.Lock()
	if data.servers[sessionID] == nil {
		data.servers[sessionID] = &ServerInfo{
			History: make([]*SystemInfo, 0, serverConfig.DataLimit),
		}
	}
	data.servers[sessionID].OwnerUserID = user.ID
	data.mu.Unlock()

	response := SessionRegisterResponse{
		SessionID: sessionID,
		Hostname:  req.Hostname,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	log.Printf("[Session注册] 成功为主机 %s 分配Session ID: %s，归属用户: %s", req.Hostname, sessionID, user.Login)
}

// handleGenerateAccessKey godoc
// @Summary 生成访问密钥
// @Tags access
// @Accept json
// @Produce json
// @Param payload body AccessKeyRequest true "双密钥"
// @Success 200 {object} AccessKeyResponse
// @Failure 400 {string} string "无效的请求格式"
// @Failure 401 {string} string "无效的服务器密钥或项目密钥"
// @Router /generate-access-key [post]
func handleGenerateAccessKey(w http.ResponseWriter, r *http.Request) {
	var req AccessKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	// 验证双密钥
	if !validateDualKey(req.ServerKey, req.ProjectKey) {
		http.Error(w, "无效的服务器密钥或项目密钥", http.StatusUnauthorized)
		return
	}

	// 生成访问密钥
	accessKey := generateAccessKey(req.ServerKey, req.ProjectKey)

	response := AccessKeyResponse{
		AccessKey:  accessKey,
		ProjectKey: req.ProjectKey,
		Message:    "访问密钥生成成功，可用于访问 " + req.ProjectKey + " 项目的面板",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleGetServersByAccessKey godoc
// @Summary 使用访问密钥获取服务器列表
// @Tags access
// @Produce json
// @Param accessKey path string true "访问密钥"
// @Success 200 {array} ServerStatus
// @Failure 401 {string} string "无效的访问密钥"
// @Router /access/{accessKey}/servers [get]
func handleGetServersByAccessKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accessKey := vars["accessKey"]

	// 验证访问密钥格式
	if accessKey == "" {
		http.Error(w, "无效的访问密钥", http.StatusUnauthorized)
		return
	}

	data.mu.RLock()
	defer data.mu.RUnlock()

	now := time.Now()
	var servers []ServerStatus

	// 遍历所有服务器，查找匹配访问密钥的数据
	for _, server := range data.servers {
		if server.Latest == nil {
			continue
		}

		// 检查服务器数据是否匹配访问密钥
		// 这里需要根据服务器的ProjectKey重新计算访问密钥进行匹配
		if !isServerMatchingAccessKey(server.Latest.ProjectKey, accessKey) {
			continue
		}

		status := "online"
		if now.Sub(server.Latest.Timestamp) > offlineThreshold {
			status = "offline"
		}

		servers = append(servers, ServerStatus{
			Hostname:         server.Latest.Hostname,
			SessionID:        server.Latest.SessionID,
			LastSeen:         server.LastSeen,
			Status:           status,
			CPUPercent:       server.Latest.CPU.UsagePercent,
			MemoryPercent:    server.Latest.Memory.UsagePercent,
			DiskPercent:      server.Latest.Disk.UsagePercent,
			OS:               server.Latest.OS.Platform,
			CPUTemp:          server.Latest.Temperature.CPUTemp,
			GPUTemp:          server.Latest.Temperature.GPUTemp,
			GPUs:             server.Latest.GPUs, // 添加所有GPU信息
			MaxTemp:          server.Latest.Temperature.MaxTemp,
			NetworkSpeedSent: server.Latest.Network.SpeedSent,
			NetworkSpeedRecv: server.Latest.Network.SpeedRecv,
			NetworkBytesSent: server.Latest.Network.BytesSent,
			NetworkBytesRecv: server.Latest.Network.BytesRecv,
			UserResources:    server.Latest.UserResources,
		})
	}

	// 按主机名排序
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(servers); err != nil {
		log.Printf("Error encoding servers response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleGetServerBySessionID godoc
// @Summary 使用访问密钥+SessionID 获取服务器详情
// @Tags access
// @Produce json
// @Param accessKey path string true "访问密钥"
// @Param sessionID path string true "Session ID"
// @Success 200 {object} ServerInfo
// @Failure 401 {string} string "无效的访问密钥"
// @Failure 404 {string} string "服务器不存在"
// @Router /access/{accessKey}/server-by-session/{sessionID} [get]
func handleGetServerBySessionID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accessKey := vars["accessKey"]
	sessionID := vars["sessionID"]

	// 验证访问密钥格式
	if accessKey == "" {
		http.Error(w, "无效的访问密钥", http.StatusUnauthorized)
		return
	}

	data.mu.RLock()
	defer data.mu.RUnlock()

	server, exists := data.servers[sessionID]
	if !exists {
		http.Error(w, "服务器不存在", http.StatusNotFound)
		return
	}

	// 检查服务器是否匹配访问密钥
	if server.Latest == nil || !isServerMatchingAccessKey(server.Latest.ProjectKey, accessKey) {
		http.Error(w, "服务器不属于指定访问密钥或无数据", http.StatusForbidden)
		return
	}

	// 过滤历史数据，只返回匹配访问密钥的数据
	filteredServer := &ServerInfo{
		History:  make([]*SystemInfo, 0),
		Latest:   server.Latest,
		LastSeen: server.LastSeen,
	}

	for _, historyItem := range server.History {
		if isServerMatchingAccessKey(historyItem.ProjectKey, accessKey) {
			filteredServer.History = append(filteredServer.History, historyItem)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(filteredServer)
}

// isServerMatchingAccessKey 检查服务器数据是否匹配访问密钥
func isServerMatchingAccessKey(serverProjectKey, accessKey string) bool {
	// 这里需要实现逻辑来检查服务器的ProjectKey是否能生成匹配的访问密钥
	// 由于我们不知道原始的Server-Key和Project-Key组合，这里使用简化的匹配逻辑
	// 实际应用中，可能需要在服务器数据中存储更多信息来支持这种匹配

	// 临时解决方案：将serverProjectKey作为projectKey，与配置的ServerKey组合生成访问密钥进行比较
	if serverConfig.ServerKey != "" {
		expectedAccessKey := generateAccessKey(serverConfig.ServerKey, serverProjectKey)
		return expectedAccessKey == accessKey
	}

	return false
}

// handleGetUserResources godoc
// @Summary 获取指定服务器的用户资源使用情况
// @Tags user_resources
// @Produce json
// @Param sessionID path string true "SessionID"
// @Success 200 {array} UserResourceInfo
// @Failure 404 {string} string "服务器不存在或没有数据"
// @Router /user-resources/{sessionID} [get]
func handleGetUserResources(w http.ResponseWriter, r *http.Request) {
	user, _ := getSessionUser(r)
	if user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["sessionID"]

	server := getServerBySessionID(sessionID)
	if server == nil {
		http.Error(w, "服务器不存在", http.StatusNotFound)
		return
	}

	if server.OwnerUserID > 0 && server.OwnerUserID != user.ID {
		http.Error(w, "无权访问该服务器", http.StatusForbidden)
		return
	}

	// 检查是否有用户资源数据
	if server.Latest == nil || len(server.Latest.UserResources) == 0 {
		http.Error(w, "没有用户资源数据", http.StatusNotFound)
		return
	}

	// 返回用户资源数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(server.Latest.UserResources); err != nil {
		log.Printf("Error encoding user resources: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleGetUserResourcesByAccessKey godoc
// @Summary 使用访问密钥获取用户资源数据
// @Tags access
// @Produce json
// @Param accessKey path string true "访问密钥"
// @Param sessionID path string true "SessionID"
// @Success 200 {array} UserResourceInfo
// @Failure 401 {string} string "无效的访问密钥"
// @Failure 404 {string} string "服务器不存在或没有数据"
// @Router /access/{accessKey}/user-resources/{sessionID} [get]
func handleGetUserResourcesByAccessKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accessKey := vars["accessKey"]
	sessionID := vars["sessionID"]

	// 验证访问密钥格式
	if accessKey == "" {
		http.Error(w, "无效的访问密钥", http.StatusUnauthorized)
		return
	}

	data.mu.RLock()
	defer data.mu.RUnlock()

	// 查找服务器
	matchedServer := getServerBySessionID(sessionID)

	if matchedServer == nil {
		http.Error(w, "服务器不存在或访问被拒绝", http.StatusNotFound)
		return
	}

	// 检查是否有用户资源数据
	if len(matchedServer.Latest.UserResources) == 0 {
		http.Error(w, "没有用户资源数据", http.StatusNotFound)
		return
	}

	// 返回用户资源数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matchedServer.Latest.UserResources); err != nil {
		log.Printf("Error encoding user resources: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleGetUUIDCount godoc
// @Summary 获取UUID数量统计
// @Tags stats
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /uuid-count [get]
func handleGetUUIDCount(w http.ResponseWriter, r *http.Request) {
	var response map[string]interface{}
	var err error
	ctx := r.Context()

	// 尝试从Redis缓存获取数据
	if serverConfig.EnableCache {
		cachedResponse, err := cacheManager.GetUUIDStats(ctx)
		if err == nil && len(cachedResponse) > 0 {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cachedResponse)
			log.Printf("从Redis缓存获取UUID统计")
			return
		}
	}

	// 检查内存缓存是否有效（1分钟内）
	data.uuidCacheMutex.RLock()
	cacheValid := time.Since(data.uuidCacheTime) < time.Minute
	if cacheValid && len(data.uuidStatsCache) > 0 {
		// 使用内存缓存数据
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data.uuidStatsCache)
		data.uuidCacheMutex.RUnlock()
		return
	}
	data.uuidCacheMutex.RUnlock()

	// 优先从数据库获取统计数据
	if data.database != nil {
		response, err = data.database.GetUUIDStats()
		if err != nil {
			log.Printf("从数据库获取UUID统计失败: %v", err)
			// 降级到内存计算
			response = calculateUUIDStatsFromMemory()
		}
	} else {
		// 从内存计算统计数据
		response = calculateUUIDStatsFromMemory()
	}

	// 更新内存缓存
	data.uuidCacheMutex.Lock()
	data.uuidStatsCache = response
	data.uuidCacheTime = time.Now()
	data.uuidCacheMutex.Unlock()

	// 更新Redis缓存
	if serverConfig.EnableCache {
		go func() {
			cacheCtx := context.Background()
			_ = cacheManager.SetUUIDStats(cacheCtx, response)
			log.Printf("UUID统计已缓存到Redis")
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// calculateUUIDStatsFromMemory 从内存计算UUID统计（降级方案）
func calculateUUIDStatsFromMemory() map[string]interface{} {
	data.mu.RLock()
	defer data.mu.RUnlock()

	// 统计活跃的UUID数量（有session ID的服务器）
	activeUUIDs := 0
	totalServers := len(data.servers)

	for _, server := range data.servers {
		// 检查是否有有效的session ID（不是hostname）
		if server.Latest != nil && server.Latest.SessionID != "" && server.Latest.SessionID != server.Latest.Hostname {
			activeUUIDs++
		}
	}

	// 构造响应
	response := map[string]interface{}{
		"total_servers": totalServers,
		"active_uuids":  activeUUIDs,
		"hostname_only": totalServers - activeUUIDs,
		"timestamp":     time.Now(),
		"description":   "使用我们服务的设备统计",
	}

	return response
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// 安全检查：只允许下载特定文件
	allowedFiles := map[string]string{
		"monitor-agent-linux": "./monitor-agent-linux",
		"install-client.sh":   "./install-client.sh",
		"install.sh":          "./install.sh",
	}

	filePath, allowed := allowedFiles[filename]
	if !allowed {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "无法打开文件", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "无法获取文件信息", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", string(rune(fileInfo.Size())))

	// 发送文件
	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("文件下载失败: %v", err)
	}

	log.Printf("文件下载: %s", filename)
}

func handleInstallScript(w http.ResponseWriter, r *http.Request) {
	// 读取安装脚本
	scriptPath := "./install-client.sh"
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		// 如果文件不存在，返回内嵌的脚本
		script = []byte(getEmbeddedInstallScript())
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "inline; filename=install-client.sh")
	if _, err := w.Write(script); err != nil {
		log.Printf("Error writing install script: %v", err)
	}

	log.Printf("安装脚本下载请求来自: %s", r.RemoteAddr)
}

func handleAPIDoc(w http.ResponseWriter, r *http.Request) {
	// 获取可执行文件目录
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
		http.Error(w, "服务器内部错误", http.StatusInternalServerError)
		return
	}
	execDir := filepath.Dir(execPath)

	// 尝试多个可能的API文档路径
	docPaths := []string{
		filepath.Join(execDir, "API.md"),
		"./API.md",
		filepath.Join(filepath.Dir(execPath), "data-server", "API.md"),
	}

	var doc []byte
	var docPath string
	for _, path := range docPaths {
		doc, err = os.ReadFile(path)
		if err == nil {
			docPath = path
			break
		}
	}

	if err != nil {
		log.Printf("API文档文件未找到，尝试的路径: %v", docPaths)
		http.Error(w, "API文档不存在", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if _, err := w.Write(doc); err != nil {
		log.Printf("Error writing API doc: %v", err)
	}

	log.Printf("API文档访问请求来自: %s (文档路径: %s)", r.RemoteAddr, docPath)
}

func getEmbeddedInstallScript() string {
	return `#!/bin/bash

# ServerStatus Monitor 客户端一键安装脚本
# 自动下载并启动监控客户端

set -e

# 配置变量
SERVER_URL="http://localhost:8080"
DOWNLOAD_URL="${SERVER_URL}/download"
CLIENT_NAME="monitor-agent-linux"
INSTALL_DIR="$HOME/gpu-monitor-client"
SERVICE_NAME="gpu-monitor-agent"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== ServerStatus Monitor 客户端一键安装脚本 ===${NC}"
echo -e "${BLUE}服务器地址: ${SERVER_URL}${NC}"
echo ""

# 检查系统
echo -e "${YELLOW}检查系统环境...${NC}"
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo -e "${RED}错误: 此脚本仅支持Linux系统${NC}"
    exit 1
fi

# 检查必要工具
for cmd in curl wget; do
    if command -v $cmd >/dev/null 2>&1; then
        DOWNLOAD_CMD=$cmd
        break
    fi
done

if [ -z "$DOWNLOAD_CMD" ]; then
    echo -e "${RED}错误: 需要安装 curl 或 wget${NC}"
    echo "Ubuntu/Debian: sudo apt install curl"
    echo "CentOS/RHEL: sudo yum install curl"
    exit 1
fi

# 创建安装目录
echo -e "${YELLOW}创建安装目录: ${INSTALL_DIR}${NC}"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# 下载客户端
echo -e "${YELLOW}下载客户端程序...${NC}"
if [ "$DOWNLOAD_CMD" = "curl" ]; then
    if ! curl -L -o "$CLIENT_NAME" "${DOWNLOAD_URL}/${CLIENT_NAME}"; then
        echo -e "${RED}下载失败${NC}"
        exit 1
    fi
else
    if ! wget -O "$CLIENT_NAME" "${DOWNLOAD_URL}/${CLIENT_NAME}"; then
        echo -e "${RED}下载失败${NC}"
        exit 1
    fi
fi

# 设置执行权限
echo -e "${YELLOW}设置执行权限...${NC}"
chmod +x "$CLIENT_NAME"

# 创建启动脚本
echo -e "${YELLOW}创建启动脚本...${NC}"
cat > start-client.sh << 'EOF'
#!/bin/bash
echo "正在启动GPU监控客户端..."
echo "监控端将向 localhost:8080 上报数据"
echo "按 Ctrl+C 停止监控"
echo ""
./monitor-agent-linux
EOF

chmod +x start-client.sh

echo -e "${GREEN}=== 安装完成 ===${NC}"
echo -e "安装目录: ${INSTALL_DIR}"
echo -e "监控界面: ${SERVER_URL}"
echo ""
echo "启动命令: ./start-client.sh"
echo "后台运行: nohup ./monitor-agent-linux > agent.log 2>&1 &"
`
}

// handleWebSocketStats godoc
// @Summary 获取WebSocket连接统计信息
// @Tags websocket
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {string} string "WebSocket未启用"
// @Router /ws-stats [get]
func handleWebSocketStats(w http.ResponseWriter, r *http.Request) {
	if !serverConfig.EnableWebSocket {
		http.Error(w, "WebSocket未启用", http.StatusServiceUnavailable)
		return
	}

	stats := webSocketManager.GetStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding WebSocket stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleHealth godoc
// @Summary 健康检查
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	healthStatus := "healthy"
	components := make(map[string]interface{})

	if data.database != nil {
		if err := data.database.Ping(); err != nil {
			healthStatus = "degraded"
			components["database"] = map[string]interface{}{
				"status":  "unhealthy",
				"message": err.Error(),
			}
		} else {
			components["database"] = map[string]interface{}{
				"status":  "healthy",
				"message": "database connection is healthy",
			}
		}
	} else {
		healthStatus = "degraded"
		components["database"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": "database not initialized",
		}
	}

	if serverConfig.EnableCache && cacheManager != nil {
		if cacheManager.enabled && cacheManager.redisConn {
			components["cache"] = map[string]interface{}{
				"status":  "healthy",
				"message": "cache service is healthy",
				"type":    "redis",
			}
		} else {
			components["cache"] = map[string]interface{}{
				"status":  "degraded",
				"message": "using memory cache fallback",
				"type":    "memory",
			}
		}
	} else {
		components["cache"] = map[string]interface{}{
			"status":  "disabled",
			"message": "cache service is disabled",
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := m.Alloc / 1024 / 1024
	memoryStatus := "healthy"
	if allocMB > 512 {
		memoryStatus = "degraded"
		if healthStatus == "healthy" {
			healthStatus = "degraded"
		}
	}

	components["memory"] = map[string]interface{}{
		"status":     memoryStatus,
		"alloc_mb":   allocMB,
		"sys_mb":     m.Sys / 1024 / 1024,
		"num_gc":     m.NumGC,
		"goroutines": runtime.NumGoroutine(),
	}

	components["system"] = map[string]interface{}{
		"status":     "healthy",
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"cpu_count":  runtime.NumCPU(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}

	healthResponse := map[string]interface{}{
		"status":     healthStatus,
		"timestamp":  time.Now().UTC(),
		"version":    buildVersion,
		"uptime":     time.Since(startTime),
		"components": components,
	}

	statusCode := http.StatusOK
	if healthStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		log.Printf("Error encoding health response: %v", err)
	}
}

// handleVersion godoc
// @Summary 版本信息
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /version [get]
func handleVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]interface{}{
		"name":       "ServerStatus Data Server",
		"version":    buildVersion,
		"build_time": buildTime,
		"go_version": runtime.Version(),
		"git_commit": buildCommit,
		"hostname":   getHostname(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
		"uptime":     time.Since(startTime).String(),
		"start_time": startTime,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(version); err != nil {
		log.Printf("Error encoding version response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleSystemStats godoc
// @Summary 系统统计信息
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /stats [get]
func handleSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	}

	// 服务器配置信息
	stats["server_config"] = map[string]interface{}{
		"host":               serverConfig.Host,
		"port":               serverConfig.Port,
		"require_auth":       serverConfig.RequireAuth,
		"data_limit":         serverConfig.DataLimit,
		"data_interval":      serverConfig.DataInterval,
		"enable_compression": serverConfig.EnableCompression,
		"enable_websocket":   serverConfig.EnableWebSocket,
		"enable_cache":       serverConfig.EnableCache,
	}

	// 数据库统计
	if data.database != nil {
		serverCount := 0
		uuidCount := 0

		// 获取服务器数量（获取第一页的统计）
		if servers, err := data.database.GetAllServers("public", 0, 1); err == nil {
			serverCount = len(servers)
		}

		// 获取UUID统计
		if uuidStats, err := data.database.GetUUIDStats(); err == nil {
			if stats, ok := uuidStats["active_uuids"]; ok {
				if count, ok := stats.(int64); ok {
					uuidCount = int(count)
				}
			}
		}

		stats["database"] = map[string]interface{}{
			"status":       "connected",
			"server_count": serverCount,
			"uuid_count":   uuidCount,
			"path":         serverConfig.DatabasePath,
		}
	} else {
		stats["database"] = map[string]interface{}{
			"status": "disconnected",
		}
	}

	// 缓存统计
	if serverConfig.EnableCache && cacheManager != nil {
		if cacheStats, err := cacheManager.GetStats(r.Context()); err == nil {
			stats["cache"] = cacheStats
		} else {
			stats["cache"] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		}
	}

	// WebSocket统计
	if serverConfig.EnableWebSocket && webSocketManager != nil {
		stats["websocket"] = map[string]interface{}{
			"enabled":     true,
			"connections": webSocketManager.GetConnectionCount(),
		}
	} else {
		stats["websocket"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// 运行时统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats["runtime"] = map[string]interface{}{
		"alloc_mb":   m.Alloc / 1024 / 1024,
		"sys_mb":     m.Sys / 1024 / 1024,
		"num_gc":     m.NumGC,
		"goroutines": runtime.NumGoroutine(),
		"go_version": runtime.Version(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding system stats response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleTrackVisit godoc
// @Summary 访客埋点上报
// @Tags visitor
// @Produce image/gif
// @Param page query string false "页面URL"
// @Param referrer query string false "来源URL"
// @Param project_key query string false "项目密钥，空则自动推断或使用默认"
// @Param session_id query string false "会话ID"
// @Router /visitor/track [get]
func handleTrackVisit(w http.ResponseWriter, r *http.Request) {
	pageURL := r.URL.Query().Get("page")
	if pageURL == "" {
		pageURL = r.Referer()
	}

	referrer := r.URL.Query().Get("referrer")
	if referrer == "" {
		referrer = r.Referer()
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sid")
	}

	// 基于Referer/页面自动推断域名
	domain := getDomainFromURL(pageURL)
	if domain == "" {
		domain = getDomainFromURL(referrer)
	}

	projectKey := r.URL.Query().Get("project_key")
	if projectKey == "" {
		// 若绑定存在则使用绑定，否则使用域名作为桶，不行再回退默认
		if data.database != nil && domain != "" {
			if pk, err := data.database.GetProjectKeyByDomain(domain); err == nil && pk != "" {
				projectKey = pk
			}
		}
		if projectKey == "" {
			if domain != "" {
				projectKey = domain
			} else {
				projectKey = "serverstatus.ltd"
			}
		}
	}

	event := &VisitorEvent{
		ProjectKey: projectKey,
		Domain:     domain,
		PageURL:    pageURL,
		Referrer:   referrer,
		UserAgent:  r.UserAgent(),
		IP:         getClientIP(r),
		SessionID:  sessionID,
		Timestamp:  time.Now(),
	}

	if data.database != nil {
		if err := data.database.SaveVisitorEvent(event); err != nil {
			log.Printf("保存访客事件失败: %v", err)
		}
	}

	writeTrackingPixel(w)
}

// handleVisitorStats godoc
// @Summary 获取访客统计
// @Tags visitor
// @Produce json
// @Param project_key query string false "项目密钥"
// @Param hours query int false "最近小时数，默认24"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string "需要登录"
// @Failure 503 {string} string "数据库未初始化"
// @Router /visitor/stats [get]
func handleVisitorStats(w http.ResponseWriter, r *http.Request) {
	if data.database == nil {
		http.Error(w, "数据库未初始化", http.StatusServiceUnavailable)
		return
	}

	projectKey := r.URL.Query().Get("project_key")
	if projectKey == "" {
		projectKey = "public"
	}

	// 非public项目需要登录
	if projectKey != "public" {
		if user, _ := getSessionUser(r); user == nil {
			http.Error(w, "需要登录", http.StatusUnauthorized)
			return
		}
	}

	hours := 24
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 24*30 {
			hours = h
		}
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	stats, err := data.database.GetVisitorStats(projectKey, since)
	if err != nil {
		log.Printf("获取访客统计失败: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("编码访客统计失败: %v", err)
	}
}

// handleVisitorAggregate godoc
// @Summary 访客分组聚合
// @Tags visitor
// @Produce json
// @Param group_by query string true "page|referrer|domain|ua"
// @Param project_key query string false "项目密钥"
// @Param hours query int false "最近小时数，默认24"
// @Param limit query int false "返回条数，默认20"
// @Success 200 {array} AggregatedVisitorItem
// @Failure 400 {string} string "缺少 group_by"
// @Failure 503 {string} string "数据库未初始化"
// @Router /visitor/aggregate [get]
func handleVisitorAggregate(w http.ResponseWriter, r *http.Request) {
	if data.database == nil {
		http.Error(w, "数据库未初始化", http.StatusServiceUnavailable)
		return
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		http.Error(w, "group_by 不能为空", http.StatusBadRequest)
		return
	}

	projectKey := r.URL.Query().Get("project_key")
	if projectKey == "" {
		projectKey = "public"
	}

	// 非public项目需要登录
	if projectKey != "public" {
		if user, _ := getSessionUser(r); user == nil {
			http.Error(w, "需要登录", http.StatusUnauthorized)
			return
		}
	}

	hours := 24
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 24*30 {
			hours = h
		}
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	items, err := data.database.GetVisitorAggregation(projectKey, groupBy, since, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("编码聚合统计失败: %v", err)
	}
}

// handleAuthGitHubLogin godoc
// @Summary GitHub OAuth 登录跳转
// @Tags auth
// @Produce json
// @Success 302 {string} string "跳转至 GitHub"
// @Failure 503 {string} string "未配置 GitHub OAuth"
// @Router /auth/github/login [get]
func handleAuthGitHubLogin(w http.ResponseWriter, r *http.Request) {
	if githubClientID == "" || githubClientSecret == "" {
		http.Error(w, "未配置 GitHub OAuth（GITHUB_CLIENT_ID/SECRET）", http.StatusServiceUnavailable)
		return
	}

	state := newOAuthState()
	cb := githubCallbackURL
	scope := "read:user user:email"

	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		url.QueryEscape(githubClientID),
		url.QueryEscape(cb),
		url.QueryEscape(scope),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleAuthGitHubCallback godoc
// @Summary GitHub OAuth 回调
// @Tags auth
// @Produce json
// @Success 302 {string} string "登录成功跳转"
// @Failure 400 {string} string "state 或 code 无效"
// @Failure 502 {string} string "GitHub 认证失败"
// @Router /auth/github/callback [get]
func handleAuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if data.database == nil {
		http.Error(w, "数据库未初始化", http.StatusServiceUnavailable)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" || !validateOAuthState(state) {
		http.Error(w, "state 或 code 无效", http.StatusBadRequest)
		return
	}

	accessToken, err := exchangeGithubCode(code)
	if err != nil {
		log.Printf("GitHub code 交换失败: %v", err)
		http.Error(w, "GitHub 认证失败", http.StatusBadGateway)
		return
	}

	ghUser, err := fetchGithubUser(accessToken)
	if err != nil {
		log.Printf("GitHub 用户信息获取失败: %v", err)
		http.Error(w, "GitHub 用户信息获取失败", http.StatusBadGateway)
		return
	}

	userID, err := data.database.SaveOrUpdateUser(&User{
		GithubID:  fmt.Sprintf("%d", ghUser.ID),
		Login:     ghUser.Login,
		Name:      ghUser.Name,
		AvatarURL: ghUser.AvatarURL,
		Email:     ghUser.Email,
	})
	if err != nil {
		log.Printf("保存用户失败: %v", err)
		http.Error(w, "用户保存失败", http.StatusInternalServerError)
		return
	}

	// 如果用户没有配置，则初始化
	if existingConfig, _ := data.database.GetUserConfig(userID); existingConfig == "" {
		_ = data.database.UpsertUserConfig(userID, "{}")
	}

	sessionToken := generateRandomHex(32)
	expires := time.Now().Add(sessionTTL)
	if err := data.database.CreateSession(userID, sessionToken, expires); err != nil {
		log.Printf("创建会话失败: %v", err)
		http.Error(w, "创建会话失败", http.StatusInternalServerError)
		return
	}

	setSessionCookie(w, sessionToken, expires)
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleAuthMe godoc
// @Summary 获取当前登录用户
// @Tags auth
// @Produce json
// @Success 200 {object} AuthUser
// @Failure 401 {string} string "未登录"
// @Failure 503 {string} string "数据库未初始化"
// @Router /auth/me [get]
func handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if data.database == nil {
		http.Error(w, "数据库未初始化", http.StatusServiceUnavailable)
		return
	}

	user, err := getSessionUser(r)
	if err != nil || user == nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}

	var cfg interface{}
	if configStr, _ := data.database.GetUserConfig(user.ID); configStr != "" {
		_ = json.Unmarshal([]byte(configStr), &cfg)
	}

	resp := AuthUser{
		ID:        user.ID,
		Login:     user.Login,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Config:    cfg,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// handleAuthLogout godoc
// @Summary 登出
// @Tags auth
// @Produce json
// @Success 204 {string} string "退出成功"
// @Router /auth/logout [post]
func handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if data.database != nil {
		if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
			_ = data.database.DeleteSession(cookie.Value)
		}
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// exchangeGithubCode 用 code 换取 access token
func exchangeGithubCode(code string) (string, error) {
	values := url.Values{}
	values.Set("client_id", githubClientID)
	values.Set("client_secret", githubClientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", githubCallbackURL)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("获取access_token失败: %s", result.Error)
	}
	return result.AccessToken, nil
}

// GithubUser GitHub返回的用户信息
type GithubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// fetchGithubUser 获取 GitHub 用户信息
func fetchGithubUser(token string) (*GithubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GithubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// 如果邮箱缺失，尝试获取邮箱列表
	if user.Email == "" {
		reqEmail, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		reqEmail.Header.Set("Authorization", "Bearer "+token)
		reqEmail.Header.Set("Accept", "application/json")
		if respEmail, err := http.DefaultClient.Do(reqEmail); err == nil {
			defer respEmail.Body.Close()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := json.NewDecoder(respEmail.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary && e.Verified && e.Email != "" {
						user.Email = e.Email
						break
					}
				}
				if user.Email == "" && len(emails) > 0 {
					user.Email = emails[0].Email
				}
			}
		}
	}

	return &user, nil
}

// handleListDomainBindings godoc
// @Summary 列出域名绑定
// @Tags visitor
// @Produce json
// @Success 200 {array} DomainBinding
// @Failure 401 {string} string "需要登录"
// @Failure 503 {string} string "数据库未初始化"
// @Router /visitor/bindings [get]
func handleListDomainBindings(w http.ResponseWriter, r *http.Request) {
	if data.database == nil {
		http.Error(w, "数据库未初始化", http.StatusServiceUnavailable)
		return
	}

	if user, _ := getSessionUser(r); user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	bindings, err := data.database.ListDomainBindings()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bindings)
}

// handleUpsertDomainBinding godoc
// @Summary 新增或更新域名绑定
// @Tags visitor
// @Accept json
// @Produce json
// @Param binding body DomainBinding true "域名绑定"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "参数错误"
// @Failure 401 {string} string "需要登录"
// @Failure 503 {string} string "数据库未初始化"
// @Router /visitor/bindings [post]
func handleUpsertDomainBinding(w http.ResponseWriter, r *http.Request) {
	if data.database == nil {
		http.Error(w, "数据库未初始化", http.StatusServiceUnavailable)
		return
	}

	if user, _ := getSessionUser(r); user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	var payload DomainBinding
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	payload.Domain = strings.ToLower(strings.TrimSpace(payload.Domain))
	payload.ProjectKey = strings.TrimSpace(payload.ProjectKey)

	if payload.Domain == "" || payload.ProjectKey == "" {
		http.Error(w, "domain 和 project_key 不能为空", http.StatusBadRequest)
		return
	}

	if err := data.database.UpsertDomainBinding(payload); err != nil {
		http.Error(w, "保存域名绑定失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "ok",
	})
}

// handleReloadConfig godoc
// @Summary 配置热重载
// @Tags config
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "重载失败"
// @Router /reload-config [post]
func handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到配置重载请求")

	// 重新加载配置文件
	if err := reloadServerConfig(); err != nil {
		log.Printf("配置重载失败: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	log.Printf("✅ 配置重载成功")

	// 广播配置更新通知
	if serverConfig.EnableWebSocket && webSocketManager != nil {
		notification := WebSocketMessage{
			Type:      "config_reload",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"message": "服务器配置已热重载",
				"time":    time.Now().UTC().Format(time.RFC3339),
			},
		}
		webSocketManager.BroadcastToProject("public", notification)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置重载成功",
		"time":    time.Now().UTC().Format(time.RFC3339),
		"config": map[string]interface{}{
			"host":             serverConfig.Host,
			"port":             serverConfig.Port,
			"require_auth":     serverConfig.RequireAuth,
			"data_limit":       serverConfig.DataLimit,
			"data_interval":    serverConfig.DataInterval,
			"enable_cache":     serverConfig.EnableCache,
			"enable_websocket": serverConfig.EnableWebSocket,
			"database_driver":  serverConfig.DatabaseDriver,
			"database_path":    serverConfig.DatabasePath,
		},
	})
}

// reloadServerConfig 重新加载服务器配置
func reloadServerConfig() error {
	// 读取当前配置文件路径
	configFile := "server-config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configFile)
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var fileConfig ServerConfig
	err = json.Unmarshal(data, &fileConfig)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置文件
	if err := validateServerConfig(&fileConfig); err != nil {
		return fmt.Errorf("配置文件验证失败: %w", err)
	}

	// 更新配置（只允许热重载的部分）
	if strings.ToLower(fileConfig.DatabaseDriver) != strings.ToLower(serverConfig.DatabaseDriver) {
		log.Printf("⚠️  数据库驱动变更需要重启服务器: %s -> %s", serverConfig.DatabaseDriver, fileConfig.DatabaseDriver)
		return fmt.Errorf("数据库驱动变更需要重启服务器才能生效")
	}
	if len(fileConfig.DatabaseConns) > 0 {
		if !reflect.DeepEqual(normalizeDBConns(serverConfig.DatabaseConns), normalizeDBConns(fileConfig.DatabaseConns)) {
			log.Printf("⚠️  数据库连接列表变更需要重启服务器")
			return fmt.Errorf("数据库连接列表变更需要重启服务器才能生效")
		}
	}

	if fileConfig.Host != "" && fileConfig.Host != serverConfig.Host {
		log.Printf("⚠️  主机地址变更需要重启服务器: %s -> %s", serverConfig.Host, fileConfig.Host)
		return fmt.Errorf("主机地址变更需要重启服务器才能生效")
	}

	if fileConfig.Port != "" && fileConfig.Port != serverConfig.Port {
		log.Printf("⚠️  端口变更需要重启服务器: %s -> %s", serverConfig.Port, fileConfig.Port)
		return fmt.Errorf("端口变更需要重启服务器才能生效")
	}

	// 可以热重载的配置项
	if !fileConfig.RequireAuth && serverConfig.RequireAuth {
		log.Printf("⚠️ 认证已强制开启，忽略配置中的禁用请求")
	} else if fileConfig.RequireAuth != serverConfig.RequireAuth {
		log.Printf("🔄 认证设置更新: %v -> %v", serverConfig.RequireAuth, fileConfig.RequireAuth)
		serverConfig.RequireAuth = fileConfig.RequireAuth
	}

	if fileConfig.DataLimit > 0 && fileConfig.DataLimit != serverConfig.DataLimit {
		log.Printf("🔄 数据限制更新: %d -> %d", serverConfig.DataLimit, fileConfig.DataLimit)
		serverConfig.DataLimit = fileConfig.DataLimit
	}

	if fileConfig.DataInterval > 0 && fileConfig.DataInterval != serverConfig.DataInterval {
		log.Printf("🔄 数据间隔更新: %d -> %d", serverConfig.DataInterval, fileConfig.DataInterval)
		serverConfig.DataInterval = fileConfig.DataInterval
	}

	if fileConfig.EnableCompression != serverConfig.EnableCompression {
		log.Printf("🔄 压缩设置更新: %v -> %v", serverConfig.EnableCompression, fileConfig.EnableCompression)
		serverConfig.EnableCompression = fileConfig.EnableCompression
	}

	if fileConfig.CompressionLevel >= 1 && fileConfig.CompressionLevel <= 9 && fileConfig.CompressionLevel != serverConfig.CompressionLevel {
		log.Printf("🔄 压缩级别更新: %d -> %d", serverConfig.CompressionLevel, fileConfig.CompressionLevel)
		serverConfig.CompressionLevel = fileConfig.CompressionLevel
	}

	// WebSocket设置变更
	if fileConfig.EnableWebSocket != serverConfig.EnableWebSocket {
		if fileConfig.EnableWebSocket && !serverConfig.EnableWebSocket {
			// 启用WebSocket
			if webSocketManager != nil {
				webSocketManager.Start()
				log.Printf("✅ WebSocket已启用")
			}
		} else if !fileConfig.EnableWebSocket && serverConfig.EnableWebSocket {
			// 禁用WebSocket（不能完全停止，但可以停止接受新连接）
			log.Printf("⚠️  WebSocket禁用需要重启服务器才能完全生效")
		}
		serverConfig.EnableWebSocket = fileConfig.EnableWebSocket
	}

	// 缓存设置变更
	if fileConfig.EnableCache != serverConfig.EnableCache {
		log.Printf("🔄 缓存设置更新: %v -> %v", serverConfig.EnableCache, fileConfig.EnableCache)
		if fileConfig.EnableCache && !serverConfig.EnableCache {
			// 启用缓存
			cacheManager = NewCacheManager(fileConfig.RedisAddr, fileConfig.RedisPassword, fileConfig.RedisDB)
		} else if !fileConfig.EnableCache && serverConfig.EnableCache {
			// 禁用缓存
			if cacheManager != nil {
				_ = cacheManager.Close()
			}
			cacheManager = NewCacheManager("", "", 0)
		}
		serverConfig.EnableCache = fileConfig.EnableCache
	}

	// 更新缓存配置
	if serverConfig.EnableCache && cacheManager != nil {
		if fileConfig.RedisAddr != "" && fileConfig.RedisAddr != serverConfig.RedisAddr {
			log.Printf("🔄 Redis地址更新: %s -> %s", serverConfig.RedisAddr, fileConfig.RedisAddr)
			if cacheManager != nil {
				_ = cacheManager.Close()
			}
			cacheManager = NewCacheManager(fileConfig.RedisAddr, fileConfig.RedisPassword, fileConfig.RedisDB)
		}
		serverConfig.RedisAddr = fileConfig.RedisAddr
		serverConfig.RedisPassword = fileConfig.RedisPassword
		serverConfig.RedisDB = fileConfig.RedisDB
	}

	// 更新密钥设置
	if fileConfig.ProjectKey != "" && fileConfig.ProjectKey != serverConfig.ProjectKey {
		log.Printf("🔄 项目密钥已更新")
		serverConfig.ProjectKey = fileConfig.ProjectKey
	}

	if fileConfig.ServerKey != "" && fileConfig.ServerKey != serverConfig.ServerKey {
		log.Printf("🔄 服务器密钥已更新")
		serverConfig.ServerKey = fileConfig.ServerKey
	}

	// 数据库路径可热更新，但仍需相应迁移/重建
	if fileConfig.DatabasePath != "" && fileConfig.DatabasePath != serverConfig.DatabasePath {
		log.Printf("🔄 数据库路径已更新 (当前进程仍使用旧连接，重启后生效): %s -> %s", serverConfig.DatabasePath, fileConfig.DatabasePath)
		serverConfig.DatabasePath = fileConfig.DatabasePath
	}

	log.Printf("✅ 配置热重载完成")
	return nil
}

// handleCacheStats godoc
// @Summary 获取缓存统计信息
// @Tags cache
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {string} string "缓存未启用"
// @Router /cache-stats [get]
func handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if !serverConfig.EnableCache {
		http.Error(w, "缓存未启用", http.StatusServiceUnavailable)
		return
	}

	stats, err := cacheManager.GetStats(r.Context())
	if err != nil {
		log.Printf("Error getting cache stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding cache stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // 降低清理频率
	defer ticker.Stop()

	for range ticker.C {
		// 清理内存中的离线服务器
		data.mu.Lock()
		now := time.Now()
		for hostname, server := range data.servers {
			if server.Latest != nil && now.Sub(server.Latest.Timestamp) > 10*time.Minute {
				// WebSocket实时推送服务器离线通知
				if serverConfig.EnableWebSocket {
					serverStatus := ServerStatus{
						Hostname:         server.Latest.Hostname,
						SessionID:        server.Latest.SessionID,
						LastSeen:         server.LastSeen,
						Status:           "offline",
						CPUPercent:       server.Latest.CPU.UsagePercent,
						MemoryPercent:    server.Latest.Memory.UsagePercent,
						DiskPercent:      server.Latest.Disk.UsagePercent,
						OS:               server.Latest.OS.Platform,
						CPUTemp:          server.Latest.Temperature.CPUTemp,
						GPUTemp:          server.Latest.Temperature.GPUTemp,
						GPUs:             server.Latest.GPUs,
						MaxTemp:          server.Latest.Temperature.MaxTemp,
						NetworkSpeedSent: server.Latest.Network.SpeedSent,
						NetworkSpeedRecv: server.Latest.Network.SpeedRecv,
						NetworkBytesSent: server.Latest.Network.BytesSent,
						NetworkBytesRecv: server.Latest.Network.BytesRecv,
						UserResources:    server.Latest.UserResources,
					}
					webSocketManager.BroadcastServerUpdate(serverStatus, "offline")
				}

				log.Printf("清理长时间离线的服务器: %s", hostname)
				delete(data.servers, hostname)
			}
		}
		data.mu.Unlock()

		// 清理数据库中的旧数据
		if data.database != nil {
			if err := data.database.CleanupOldData(7*24*time.Hour, serverConfig.DataLimit); err != nil {
				log.Printf("清理数据库旧数据失败: %v", err)
			} else {
				log.Println("数据库清理完成")
			}
		}
	}
}

// isValidProjectKey 验证项目密钥
func isValidProjectKey(key string) bool {
	if key == "" {
		return false
	}

	// 检查是否与主密钥匹配
	if serverConfig.ProjectKey != "" && key == serverConfig.ProjectKey {
		return true
	}

	// 如果启用认证但没有设置具体的项目密钥，接受所有非空密钥
	if serverConfig.RequireAuth && serverConfig.ProjectKey == "" {
		return true // 接受所有非空的项目密钥
	}

	return false
}

// getServerBySessionID 仅按 sessionID 精确查找
func getServerBySessionID(sessionID string) *ServerInfo {
	data.mu.RLock()
	defer data.mu.RUnlock()
	return data.servers[sessionID]
}

// reverseHistory 将历史记录按时间升序排列（数据库返回为倒序）
func reverseHistory(history []*SystemInfo) {
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
}

// 已移除generateAccessToken函数，只保留AccessKey相关功能

// generateUUID 生成UUID字符串
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano()) // fallback
	}
	// 设置版本号和变体
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// generateAccessKey 根据Server-Key和Project-Key生成访问密钥（带缓存）
func generateAccessKey(serverKey, projectKey string) string {
	// 构造缓存键
	cacheKey := serverKey + ":" + projectKey

	// 先检查内存缓存
	accessKeyCache.mu.RLock()
	if cachedKey, exists := accessKeyCache.cache[cacheKey]; exists {
		accessKeyCache.mu.RUnlock()
		return cachedKey
	}
	accessKeyCache.mu.RUnlock()

	// 检查数据库缓存
	if data.database != nil {
		if dbKey, err := data.database.GetAccessKeyCache(cacheKey); err == nil && dbKey != "" {
			// 将数据库缓存更新到内存缓存
			accessKeyCache.mu.Lock()
			accessKeyCache.cache[cacheKey] = dbKey
			accessKeyCache.mu.Unlock()
			return dbKey
		}
	}

	// 缓存中不存在，计算新的AccessKey
	combinedKey := serverKey + ":" + projectKey
	hash := sha256.Sum256([]byte(combinedKey + "serverstatus-access-key-salt"))
	accessKey := hex.EncodeToString(hash[:])

	// 存入内存缓存
	accessKeyCache.mu.Lock()
	accessKeyCache.cache[cacheKey] = accessKey
	accessKeyCache.mu.Unlock()

	// 存入数据库缓存
	if data.database != nil {
		if err := data.database.SaveAccessKeyCache(cacheKey, accessKey); err != nil {
			log.Printf("保存访问密钥缓存到数据库失败: %v", err)
		}
	}

	return accessKey
}

// validateDualKey 验证Server-Key（Project-Key用户自定义，不验证）
func validateDualKey(serverKey, projectKey string) bool {
	if serverKey == "" {
		return false
	}

	// 只检查Server-Key是否匹配服务器配置的服务器密钥
	if serverConfig.ServerKey != "" && serverKey == serverConfig.ServerKey {
		return true
	}

	return false
}

// writeTrackingPixel 返回1x1透明GIF作为埋点响应
func writeTrackingPixel(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(trackingPixel); err != nil {
		log.Printf("返回追踪像素失败: %v", err)
	}
}

// 已移除getProjectKeyByToken函数，只保留AccessKey相关功能

// loadServerConfig 加载服务器配置文件
func loadServerConfig() {
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置文件
		saveServerConfig()
		log.Printf("创建默认服务器配置文件: %s", *configFile)
		return
	}

	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Printf("读取服务器配置文件失败: %v", err)
		return
	}

	var fileConfig ServerConfig
	err = json.Unmarshal(data, &fileConfig)
	if err != nil {
		log.Printf("❌ 解析服务器配置文件失败: %v", err)
		log.Printf("💡 请检查配置文件JSON格式是否正确")
		return
	}

	// 验证配置文件格式
	if err := validateServerConfig(&fileConfig); err != nil {
		log.Printf("❌ 服务器配置文件验证失败: %v", err)
		log.Printf("💡 请检查配置文件内容，修复后重启服务")
		return
	}

	log.Printf("✅ 配置文件验证通过")

	// 更新配置
	if fileConfig.ProjectKey != "" {
		serverConfig.ProjectKey = fileConfig.ProjectKey
	}

	if fileConfig.ServerKey != "" {
		serverConfig.ServerKey = fileConfig.ServerKey
	}

	if fileConfig.Host != "" {
		serverConfig.Host = fileConfig.Host
	}
	if fileConfig.Port != "" {
		serverConfig.Port = fileConfig.Port
	}
	// 强制启用认证，忽略配置文件中的禁用选项
	if !fileConfig.RequireAuth && serverConfig.RequireAuth {
		log.Printf("⚠️ 配置文件试图关闭认证，已被强制忽略，认证保持开启")
	} else {
		serverConfig.RequireAuth = fileConfig.RequireAuth
	}
	if fileConfig.DatabaseDriver != "" {
		serverConfig.DatabaseDriver = strings.ToLower(fileConfig.DatabaseDriver)
	}
	if fileConfig.DatabasePath != "" {
		serverConfig.DatabasePath = fileConfig.DatabasePath
	}
	if len(fileConfig.DatabaseConns) > 0 {
		// 直接覆盖列表，默认使用第一个
		serverConfig.DatabaseConns = fileConfig.DatabaseConns
	}
	if fileConfig.DataLimit > 0 {
		serverConfig.DataLimit = fileConfig.DataLimit
	}
	if fileConfig.DataInterval > 0 {
		serverConfig.DataInterval = fileConfig.DataInterval
	}
	serverConfig.EnableWebSocket = fileConfig.EnableWebSocket
	serverConfig.EnableCache = fileConfig.EnableCache
	if fileConfig.RedisAddr != "" {
		serverConfig.RedisAddr = fileConfig.RedisAddr
	}
	serverConfig.RedisPassword = fileConfig.RedisPassword
	if fileConfig.RedisDB > 0 {
		serverConfig.RedisDB = fileConfig.RedisDB
	}

	log.Printf("加载服务器配置文件: %s", *configFile)
}

// saveServerConfig 保存服务器配置文件
func saveServerConfig() {
	data, err := json.MarshalIndent(serverConfig, "", "  ")
	if err != nil {
		log.Printf("序列化服务器配置失败: %v", err)
		return
	}

	// 确保目录存在
	dir := filepath.Dir(*configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("创建服务器配置目录失败: %v", err)
		return
	}

	err = os.WriteFile(*configFile, data, 0644)
	if err != nil {
		log.Printf("保存服务器配置文件失败: %v", err)
	}
}

// initDBStore 根据配置初始化数据库存储，预留未来多驱动扩展
func initDBStore() (DBStore, error) {
	driver, path := getActiveDBConfig()
	switch strings.ToLower(driver) {
	case "", "sqlite", "sqlite3":
		return NewDatabase(path)
	case "postgres", "postgresql", "pg", "psql":
		return NewPostgresDatabase(path)
	case "mysql", "mariadb":
		return nil, fmt.Errorf("数据库驱动 mysql/mariadb 尚未实现，当前仅支持 sqlite")
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}
}

// getActiveDBConfig 选择当前使用的数据库配置（优先列表第一个）
func getActiveDBConfig() (driver string, path string) {
	if len(serverConfig.DatabaseConns) > 0 {
		first := serverConfig.DatabaseConns[0]
		driver = first.Driver
		path = first.Path
	}
	if driver == "" {
		driver = serverConfig.DatabaseDriver
	}
	if path == "" {
		path = serverConfig.DatabasePath
	}
	if driver == "" {
		driver = "sqlite"
	}
	if path == "" {
		path = "./data/serverstatus.db"
	}
	return strings.ToLower(driver), path
}

// normalizeDBConns 规范化数据库连接配置用于比较
func normalizeDBConns(conns []DBConnConfig) []DBConnConfig {
	result := make([]DBConnConfig, 0, len(conns))
	for _, c := range conns {
		result = append(result, DBConnConfig{
			Driver: strings.ToLower(c.Driver),
			Path:   c.Path,
		})
	}
	return result
}

func isSupportedDriver(driver string) bool {
	switch strings.ToLower(driver) {
	case "", "sqlite", "sqlite3", "postgres", "postgresql", "pg", "psql", "mysql", "mariadb":
		return true
	default:
		return false
	}
}

// getProjectKeyFromAccessKey 从访问密钥获取项目密钥
func getProjectKeyFromAccessKey(accessKey string) string {
	// 从数据库查询访问密钥对应的项目密钥
	projectKey, err := data.database.GetAccessKeyCache(accessKey)
	if err != nil {
		log.Printf("获取访问密钥对应的项目密钥失败: %v", err)
		return ""
	}
	return projectKey
}

// handleExportServersCSV godoc
// @Summary 导出服务器列表为CSV
// @Tags export
// @Produce text/csv
// @Param project_key query string false "项目密钥"
// @Param access_key query string false "访问密钥（用于换取项目密钥）"
// @Success 200 {file} file
// @Router /export/servers [get]
func handleExportServersCSV(w http.ResponseWriter, r *http.Request) {
	// 验证请求参数
	projectKey := r.URL.Query().Get("project_key")
	if projectKey == "" {
		// 如果没有项目密钥，尝试使用访问密钥
		accessKey := r.URL.Query().Get("access_key")
		if accessKey != "" {
			projectKey = getProjectKeyFromAccessKey(accessKey)
		}
		// 如果仍然没有项目密钥，使用默认值
		if projectKey == "" {
			projectKey = "public"
		}
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=servers_%s.csv", time.Now().Format("20060102_150405")))

	// 创建CSV写入器
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入表头
	headers := []string{
		"主机名", "会话ID", "项目密钥", "最后更新时间", "在线状态",
		"CPU使用率(%)", "CPU核心数", "CPU型号",
		"内存使用率(%)", "磁盘使用率(%)", "操作系统", "CPU温度(°C)",
	}
	if err := writer.Write(headers); err != nil {
		http.Error(w, "写入CSV失败", http.StatusInternalServerError)
		return
	}

	// 获取服务器列表数据 (使用分页)
	servers, err := data.database.GetAllServers(projectKey, 0, serverConfig.DataLimit)
	if err != nil {
		http.Error(w, "获取服务器数据失败", http.StatusInternalServerError)
		return
	}

	// 写入数据行
	for _, server := range servers {
		// 使用ServerInfo结构体，从Latest字段获取系统信息
		var hostname, sessionID, projectKey, osInfo, cpuModel string
		var cpuUsage, memoryUsage, diskUsage, cpuTemp float64
		var cpuCores int
		var lastSeen time.Time

		if server.Latest != nil {
			hostname = server.Latest.Hostname
			sessionID = server.Latest.SessionID
			projectKey = server.Latest.ProjectKey
			lastSeen = server.LastSeen
			cpuUsage = server.Latest.CPU.UsagePercent
			cpuCores = server.Latest.CPU.CoreCount
			cpuModel = server.Latest.CPU.ModelName
			memoryUsage = server.Latest.Memory.UsagePercent
			diskUsage = server.Latest.Disk.UsagePercent
			osInfo = server.Latest.OS.Platform
			cpuTemp = server.Latest.Temperature.CPUTemp
		}

		row := []string{
			hostname,
			sessionID,
			projectKey,
			lastSeen.Format("2006-01-02 15:04:05"),
			func() string {
				if time.Since(lastSeen) < time.Duration(serverConfig.DataInterval)*2*time.Second {
					return "在线"
				}
				return "离线"
			}(),
			fmt.Sprintf("%.2f", cpuUsage),
			strconv.Itoa(cpuCores),
			cpuModel,
			fmt.Sprintf("%.2f", memoryUsage),
			fmt.Sprintf("%.2f", diskUsage),
			osInfo,
			fmt.Sprintf("%.1f", cpuTemp),
		}

		if err := writer.Write(row); err != nil {
			http.Error(w, "写入CSV数据失败", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("[数据导出] 导出服务器列表CSV，项目: %s, 记录数: %d", projectKey, len(servers))
}

// handleExportHistoryCSV godoc
// @Summary 导出服务器历史数据为CSV
// @Tags export
// @Produce text/csv
// @Param session_id query string true "SessionID"
// @Param project_key query string false "项目密钥"
// @Param limit query int false "最多返回条数，默认DataLimit"
// @Param hours query int false "回溯小时数，默认24"
// @Success 200 {file} file
// @Failure 400 {string} string "缺少session_id"
// @Router /export/history [get]
func handleExportHistoryCSV(w http.ResponseWriter, r *http.Request) {
	user, _ := getSessionUser(r)
	if user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "需要提供session_id参数", http.StatusBadRequest)
		return
	}

	projectKey := r.URL.Query().Get("project_key")
	// 默认允许 public/缺省

	limit := serverConfig.DataLimit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 10000 {
			limit = l
		}
	}

	// 可选时间范围
	to := time.Now()
	from := to.Add(-24 * time.Hour)
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 24*30 {
			from = to.Add(-time.Duration(h) * time.Hour)
		}
	}

	server := getServerBySessionID(sessionID)
	if server == nil {
		http.Error(w, "服务器不存在", http.StatusNotFound)
		return
	}
	if server.OwnerUserID > 0 && server.OwnerUserID != user.ID {
		http.Error(w, "无权访问该服务器", http.StatusForbidden)
		return
	}

	// 从数据库获取历史数据
	history, err := data.database.GetHistoryByTimeRange(sessionID, projectKey, from, to, limit)
	if err != nil {
		log.Printf("获取历史数据失败: %v", err)
		http.Error(w, "获取历史数据失败", http.StatusInternalServerError)
		return
	}
	// 按时间升序
	reverseHistory(history)

	// 设置响应头
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=history_%s_%s.csv", sessionID, time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	headers := []string{
		"时间戳", "主机名", "SessionID", "项目密钥",
		"CPU使用率(%)", "内存使用率(%)", "磁盘使用率(%)",
		"网络发送字节", "网络接收字节",
		"CPU温度", "GPU温度",
	}
	if err := writer.Write(headers); err != nil {
		http.Error(w, "写入CSV失败", http.StatusInternalServerError)
		return
	}

	for _, item := range history {
		row := []string{
			item.Timestamp.Format("2006-01-02 15:04:05"),
			item.Hostname,
			item.SessionID,
			item.ProjectKey,
			fmt.Sprintf("%.2f", item.CPU.UsagePercent),
			fmt.Sprintf("%.2f", item.Memory.UsagePercent),
			fmt.Sprintf("%.2f", item.Disk.UsagePercent),
			fmt.Sprintf("%d", item.Network.BytesSent),
			fmt.Sprintf("%d", item.Network.BytesRecv),
			fmt.Sprintf("%.1f", item.Temperature.CPUTemp),
			fmt.Sprintf("%.1f", item.Temperature.GPUTemp),
		}
		if err := writer.Write(row); err != nil {
			http.Error(w, "写入CSV数据失败", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("[数据导出] 导出历史数据CSV，Session: %s, 条数: %d", sessionID, len(history))
}

// handleExportUserResourcesCSV godoc
// @Summary 导出指定服务器的用户资源为CSV
// @Tags export
// @Produce text/csv
// @Param session_id query string true "SessionID"
// @Param project_key query string false "项目密钥"
// @Success 200 {file} file
// @Failure 400 {string} string "缺少session_id"
// @Router /export/user-resources [get]
func handleExportUserResourcesCSV(w http.ResponseWriter, r *http.Request) {
	user, _ := getSessionUser(r)
	if user == nil {
		http.Error(w, "需要登录", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "需要提供session_id参数", http.StatusBadRequest)
		return
	}

	// 使用当前最新的数据导出
	server := getServerBySessionID(sessionID)
	if server == nil || server.Latest == nil || len(server.Latest.UserResources) == 0 {
		http.Error(w, "没有用户资源数据", http.StatusNotFound)
		return
	}

	if server.OwnerUserID > 0 && server.OwnerUserID != user.ID {
		http.Error(w, "无权访问该服务器", http.StatusForbidden)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=user_resources_%s_%s.csv", sessionID, time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	headers := []string{
		"时间戳", "主机名", "SessionID", "用户", "UID",
		"进程数", "CPU使用率(%)", "内存(MB)", "内存占比(%)",
	}
	if err := writer.Write(headers); err != nil {
		http.Error(w, "写入CSV失败", http.StatusInternalServerError)
		return
	}

	for _, ur := range server.Latest.UserResources {
		row := []string{
			server.Latest.Timestamp.Format("2006-01-02 15:04:05"),
			server.Latest.Hostname,
			server.Latest.SessionID,
			ur.Username,
			fmt.Sprintf("%d", ur.UID),
			fmt.Sprintf("%d", ur.ProcessCount),
			fmt.Sprintf("%.2f", ur.CPUPercent),
			fmt.Sprintf("%d", ur.MemoryMB),
			fmt.Sprintf("%.2f", ur.MemoryPercent),
		}
		if err := writer.Write(row); err != nil {
			http.Error(w, "写入CSV数据失败", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("[数据导出] 导出用户资源CSV，Session: %s, 条数: %d", sessionID, len(server.Latest.UserResources))
}

// handleDeployAgentCommand godoc
// @Summary 获取一键运行客户端的命令与示例.env
// @Tags deploy
// @Produce json
// @Param base_url query string false "数据上报基址，默认 https://serverstatus.ltd"
// @Param user_resources query bool false "是否开启用户资源采集，默认false"
// @Success 200 {object} map[string]interface{}
// @Router /deploy/agent-command [get]
func handleDeployAgentCommand(w http.ResponseWriter, r *http.Request) {
	baseURL := r.URL.Query().Get("base_url")
	if baseURL == "" {
		baseURL = "https://serverstatus.ltd"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	serverURL := fmt.Sprintf("%s/api/data", baseURL)

	enableUserResources := false
	if v := r.URL.Query().Get("user_resources"); strings.EqualFold(v, "true") || v == "1" {
		enableUserResources = true
	}

	linuxCmd := fmt.Sprintf(
		"./monitor-agent -url \"%s\" -key \"%s\" -server-key \"%s\" -user-resources=%t",
		serverURL, serverConfig.ProjectKey, serverConfig.ServerKey, enableUserResources,
	)

	// 默认假设 Linux amd64 二进制名称，可在命令中通过 AGENT_BIN 覆盖
	agentBinary := "monitor-agent-linux-amd64"
	downloadURL := fmt.Sprintf("https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/%s", agentBinary)

	linuxInstall := fmt.Sprintf(
		"AGENT_BIN=${AGENT_BIN:-%s} SERVER_URL=\"%s\" PROJECT_KEY=\"%s\" SERVER_KEY=\"%s\" ENABLE_USER_RESOURCES=%t; "+
			"curl -L -o \"$AGENT_BIN\" \"%s\" && chmod +x \"$AGENT_BIN\" && ./\"$AGENT_BIN\" -url \"$SERVER_URL\" -key \"$PROJECT_KEY\" -server-key \"$SERVER_KEY\" -user-resources=$ENABLE_USER_RESOURCES",
		agentBinary, serverURL, serverConfig.ProjectKey, serverConfig.ServerKey, enableUserResources, downloadURL,
	)

	windowsCmd := fmt.Sprintf(
		"monitor-agent.exe -url \"%%SERVER_URL%%\" -key \"%%PROJECT_KEY%%\" -server-key \"%%SERVER_KEY%%\" -user-resources=%%ENABLE_USER_RESOURCES%%",
	)

	windowsInstall := fmt.Sprintf(
		"$env:SERVER_URL=\"%s\"; $env:PROJECT_KEY=\"%s\"; $env:SERVER_KEY=\"%s\"; $env:ENABLE_USER_RESOURCES=%t; "+
			"$bin=\"monitor-agent-windows-amd64.exe\"; "+
			"Invoke-WebRequest -Uri \"https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/$bin\" -OutFile $bin; "+
			"./$bin -url $env:SERVER_URL -key $env:PROJECT_KEY -server-key $env:SERVER_KEY -user-resources=$env:ENABLE_USER_RESOURCES",
		serverURL, serverConfig.ProjectKey, serverConfig.ServerKey, enableUserResources,
	)

	envContent := fmt.Sprintf(
		"SERVER_URL=%s\nPROJECT_KEY=%s\nSERVER_KEY=%s\nENABLE_USER_RESOURCES=%t\n",
		serverURL, serverConfig.ProjectKey, serverConfig.ServerKey, enableUserResources,
	)

	response := map[string]interface{}{
		"server_url":      serverURL,
		"project_key":     serverConfig.ProjectKey,
		"server_key":      serverConfig.ServerKey,
		"linux_command":   linuxCmd,
		"windows_command": windowsCmd,
		"linux_install":   linuxInstall,
		"windows_install": windowsInstall,
		"env":             envContent,
		"download_urls": map[string]string{
			"latest_release": "https://github.com/MyDailyCloud/ServerStatus/releases/latest",
		},
		"note": "已提供直接下载并运行的一键命令，若需自定义架构请调整 AGENT_BIN 或下载 URL。",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding deploy command response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// printServerUsage 打印服务器使用说明
func printServerUsage() {
	fmt.Println("ServerStatus Monitor Data Server - 监控数据服务器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  data-server [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -key string")
	fmt.Println("        项目密钥 (用于生成访问令牌和访问密钥计算)")
	fmt.Println("  -server-key string")
	fmt.Println("        服务器密钥 (用于双密钥认证)")
	fmt.Println("  -host string")
	fmt.Println("        服务器绑定IP地址 (默认: 0.0.0.0)")
	fmt.Println("  -port string")
	fmt.Println("        服务器端口 (默认: 8080)")
	fmt.Println("  -config string")
	fmt.Println("        服务器配置文件路径 (默认: server-config.json)")
	fmt.Println("  -auth")
	fmt.Println("        启用API密钥认证")
	fmt.Println("  -data-limit int")
	fmt.Println("        每台客户端数据保留条数限制 (默认: 1000)")
	fmt.Println("  -data-interval int")
	fmt.Println("        推荐的数据上报间隔秒数 (默认: 5)")
	fmt.Println("  -websocket")
	fmt.Println("        启用WebSocket实时推送 (默认: true)")
	fmt.Println("  -cache")
	fmt.Println("        启用Redis缓存 (默认: true)")
	fmt.Println("  -redis")
	fmt.Println("        Redis服务器地址 (默认: localhost:6379)")
	fmt.Println("  -redis-password")
	fmt.Println("        Redis服务器密码 (默认: 无)")
	fmt.Println("  -redis-db")
	fmt.Println("        Redis数据库编号 (默认: 0)")
	fmt.Println("  -db-driver string")
	fmt.Println("        数据库驱动 (sqlite; 未来: postgres/mysql)")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 使用默认配置启动")
	fmt.Println("  data-server")
	fmt.Println()
	fmt.Println("  # 启用认证并设置项目密钥")
	fmt.Println("  data-server -auth -key your-project-key")
	fmt.Println()
	fmt.Println("  # 启用双密钥认证")
	fmt.Println("  data-server -server-key server-secret-key")
	fmt.Println()
	fmt.Println("  # 自定义端口")
	fmt.Println("  data-server -port 9090")
	fmt.Println()
	fmt.Println("  # 绑定到特定IP地址")
	fmt.Println("  data-server -host 192.168.1.100 -port 8080")
	fmt.Println()
	fmt.Println("  # 只监听本地连接")
	fmt.Println("  data-server -host 127.0.0.1")
	fmt.Println()
	fmt.Println("  # 使用配置文件")
	fmt.Println("  data-server -config /path/to/server-config.json")
	fmt.Println()
	fmt.Println("配置文件格式 (JSON):")
	fmt.Println(`  {`)
	fmt.Println(`    "project_key": "project-secret-key",`)
	fmt.Println(`    "allowed_keys": [`)
	fmt.Println(`      "key1",`)
	fmt.Println(`      "key2",`)
	fmt.Println(`      "key3"`)
	fmt.Println(`    ],`)
	fmt.Println(`    "server_key": "server-secret-key",`)
	fmt.Println(`    "host": "0.0.0.0",`)
	fmt.Println(`    "port": "8080",`)
	fmt.Println(`    "require_auth": true,`)
	fmt.Println(`    "data_limit": 1000,`)
	fmt.Println(`    "data_interval": 5`)
	fmt.Println(`  }`)
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  POST /api/data       - 接收监控数据上报")
	fmt.Println("  POST /api/register-session - 注册新的session获取UUID")
	fmt.Println("  GET  /api/servers    - 获取服务器列表")
	fmt.Println("  GET  /api/server/{sessionID} - 获取特定服务器详情")
	fmt.Println("  POST /api/generate-access-key - 生成访问密钥 (双密钥认证)")
	fmt.Println("  GET  /api/access/{accessKey}/servers - 根据访问密钥获取服务器列表")
	fmt.Println("  GET  /api/access/{accessKey}/server-by-session/{sessionID} - 根据访问密钥和sessionID获取特定服务器")
	fmt.Println("  GET  /api/ws-stats - 获取WebSocket连接统计信息")
	fmt.Println("  GET  /api/cache-stats - 获取缓存统计信息")
	fmt.Println("  GET  /ws - WebSocket实时数据推送连接")

	fmt.Println()
	fmt.Println("双密钥认证使用说明:")
	fmt.Println("  1. 服务器端设置服务器密钥:")
	fmt.Println("     data-server -server-key server-secret-key")
	fmt.Println()
	fmt.Println("  2. 客户端使用服务器密钥+项目密钥生成访问密钥:")
	fmt.Println("     curl -X POST http://server:8080/api/generate-access-key \\")
	fmt.Println("          -H \"Content-Type: application/json\" \\")
	fmt.Println("          -d '{\"server_key\": \"server-secret-key\", \"project_key\": \"project-alpha\"}'")
	fmt.Println()
	fmt.Println("  3. 使用访问密钥访问团队项目面板:")
	fmt.Println("     curl http://server:8080/api/access/{accessKey}/servers")
	fmt.Println()
	fmt.Println("  双密钥认证的优势:")
	fmt.Println("  - 服务器端只需设置一个服务器密钥")
	fmt.Println("  - 客户端通过服务器密钥+项目密钥组合认证")
	fmt.Println("  - 生成的访问密钥可用于访问对应项目的监控面板")
	fmt.Println("  - 不同项目使用不同的项目密钥，数据隔离")
	fmt.Println("  - 访问密钥基于SHA256哈希，安全可靠")
	fmt.Println()
	fmt.Println("WebSocket实时推送:")
	fmt.Println("  连接地址: ws://server:port/ws?project_key=your-key")
	fmt.Println("  消息格式: JSON")
	fmt.Println("  消息类型:")
	fmt.Println("    - server_update: 服务器状态更新")
	fmt.Println("    - heartbeat: 心跳消息")
	fmt.Println("  示例前端代码:")
	fmt.Println("    const ws = new WebSocket('ws://localhost:8080/ws');")
	fmt.Println("    ws.onmessage = (event) => {")
	fmt.Println("      const data = JSON.parse(event.data);")
	fmt.Println("      console.log('实时更新:', data);")
	fmt.Println("    };")
	fmt.Println()
	fmt.Println("WebSocket的优势:")
	fmt.Println("  - 实时数据推送，无需轮询")
	fmt.Println("  - 减少服务器负载和网络带宽")
	fmt.Println("  - 即时响应服务器状态变化")
	fmt.Println("  - 支持在线/离线状态通知")
	fmt.Println()
	fmt.Println("Redis缓存:")
	fmt.Println("  用途: 缓存频繁访问的数据，提升API响应速度")
	fmt.Println("  缓存内容:")
	fmt.Println("    - 服务器列表 (30秒TTL)")
	fmt.Println("    - UUID统计数据 (60秒TTL)")
	fmt.Println("    - 服务器总数 (30秒TTL)")
	fmt.Println("  启用方式:")
	fmt.Println("    data-server -cache -redis localhost:6379")
	fmt.Println("  缓存优势:")
	fmt.Println("    - API响应时间减少 50-70%")
	fmt.Println("    - 数据库查询压力大幅降低")
	fmt.Println("    - 支持缓存统计和监控")
	fmt.Println("    - 自动降级到内存缓存")
	fmt.Println()
	fmt.Println("Redis配置说明:")
	fmt.Println("  - 默认地址: localhost:6379")
	fmt.Println("  - 支持密码认证和数据库选择")
	fmt.Println("  - 连接失败时自动降级到内存模式")
	fmt.Println("  - 缓存键自动过期和清理")
}
