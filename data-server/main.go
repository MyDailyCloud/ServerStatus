package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/NYTimes/gziphandler"
)

type SystemInfo struct {
	Hostname      string              `json:"hostname"`
	SessionID     string              `json:"session_id,omitempty"` // UUID session标识
	Timestamp     time.Time           `json:"timestamp"`
	CPU           CPUInfo             `json:"cpu"`
	Memory        MemInfo             `json:"memory"`
	Disk          DiskInfo            `json:"disk"`
	Network       NetInfo             `json:"network"`
	GPU           GPUInfo             `json:"gpu"`  // 保持兼容性，主GPU信息
	GPUs          []GPUInfo           `json:"gpus"` // 所有GPU信息
	OS            OSInfo              `json:"os"`
	Temperature   TempInfo            `json:"temperature"`
	ProjectKey    string              `json:"project_key,omitempty"`
	UserResources []UserResourceInfo  `json:"user_resources,omitempty"` // 用户资源使用信息
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
	BytesSent    uint64        `json:"bytes_sent"`     // 总发送字节数
	BytesRecv    uint64        `json:"bytes_recv"`     // 总接收字节数
	PacketsSent  uint64        `json:"packets_sent"`   // 总发送包数
	PacketsRecv  uint64        `json:"packets_recv"`   // 总接收包数
	SpeedSent    float64       `json:"speed_sent"`     // 发送速率 (KB/s)
	SpeedRecv    float64       `json:"speed_recv"`     // 接收速率 (KB/s)
	Interfaces   []NetInterface `json:"interfaces"`     // 网卡详细信息
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
	Platform string `json:"platform"`
	Version  string `json:"version"`
	Arch     string `json:"arch"`
	Uptime   uint64 `json:"uptime"`
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
	database       *Database              // 数据库实例
}

type ServerInfo struct {
	Latest   *SystemInfo   `json:"latest"`
	History  []*SystemInfo `json:"history"`
	LastSeen time.Time     `json:"last_seen"`
}

type ServerStatus struct {
	Hostname          string              `json:"hostname"`
	SessionID         string              `json:"session_id,omitempty"` // UUID session标识
	LastSeen          time.Time           `json:"last_seen"`
	Status            string              `json:"status"`
	CPUPercent        float64             `json:"cpu_percent"`
	MemoryPercent     float64             `json:"memory_percent"`
	DiskPercent       float64             `json:"disk_percent"`
	OS                string              `json:"os"`
	CPUTemp           float64             `json:"cpu_temp"`
	GPUTemp           float64             `json:"gpu_temp"` // 保持兼容性，主GPU温度
	GPUs              []GPUInfo           `json:"gpus"`     // 所有GPU信息
	MaxTemp           float64             `json:"max_temp"`
	NetworkSpeedSent  float64             `json:"network_speed_sent"`  // 网络发送速率 (KB/s)
	NetworkSpeedRecv  float64             `json:"network_speed_recv"`  // 网络接收速率 (KB/s)
	NetworkBytesSent  uint64              `json:"network_bytes_sent"`  // 总发送字节数
	NetworkBytesRecv  uint64              `json:"network_bytes_recv"`  // 总接收字节数
	UserResources     []UserResourceInfo  `json:"user_resources,omitempty"` // 用户资源使用信息
}

type ServerConfig struct {
	ProjectKey        string `json:"project_key"`
	ServerKey         string `json:"server_key"`
	Host              string `json:"host"`
	Port              string `json:"port"`
	RequireAuth       bool   `json:"require_auth"`
	DataLimit         int    `json:"data_limit"`      // 数据保留条数限制
	DataInterval      int    `json:"data_interval"`   // 数据上报间隔(秒)
	DatabasePath      string `json:"database_path"`   // 数据库文件路径
	EnableCompression bool   `json:"enable_compression"` // 启用gzip压缩
	CompressionLevel  int    `json:"compression_level"`   // 压缩级别(1-9)
	EnableWebSocket   bool   `json:"enable_websocket"`   // 启用WebSocket实时推送
	EnableCache       bool   `json:"enable_cache"`       // 启用Redis缓存
	RedisAddr         string `json:"redis_addr"`         // Redis地址
	RedisPassword     string `json:"redis_password"`     // Redis密码
	RedisDB           int    `json:"redis_db"`           // Redis数据库编号
}

// AccessKey缓存结构
type AccessKeyCache struct {
	mu    sync.RWMutex
	cache map[string]string // key: serverKey:projectKey, value: accessKey
}

var (
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
		RequireAuth:       false,
		DataLimit:         1000, // 默认保留1000条数据
		DataInterval:      5,    // 默认5秒间隔
		DatabasePath:      "./data/serverstatus.db", // 数据库路径
		EnableCompression: true, // 默认启用压缩
		CompressionLevel:  6,    // 默认压缩级别
		EnableWebSocket:   true, // 默认启用WebSocket
		EnableCache:       true, // 默认启用缓存
		RedisAddr:         "localhost:6379", // 默认Redis地址
		RedisPassword:     "",    // 默认无密码
		RedisDB:           0,     // 默认数据库
	}

	// 全局AccessKey缓存
	accessKeyCache = &AccessKeyCache{
		cache: make(map[string]string),
	}

	// 命令行参数
	projectKey         = flag.String("key", "", "项目认证密钥")
	serverKey          = flag.String("server-key", "", "服务器密钥 (用于双密钥认证)")
	host               = flag.String("host", "0.0.0.0", "服务器绑定IP地址")
	port               = flag.String("port", "8080", "服务器端口")
	configFile         = flag.String("config", "server-config.json", "服务器配置文件路径")
	requireAuth        = flag.Bool("auth", false, "是否要求API密钥认证")
	dataLimit          = flag.Int("data-limit", 1000, "数据保留条数限制")
	dataInterval       = flag.Int("data-interval", 5, "推荐的数据上报间隔(秒)")
	enableCompression  = flag.Bool("compression", true, "启用gzip压缩")
	compressionLevel   = flag.Int("compression-level", 6, "gzip压缩级别(1-9)")
	enableWebSocket    = flag.Bool("websocket", true, "启用WebSocket实时推送")
	enableCache        = flag.Bool("cache", true, "启用Redis缓存")
	redisAddr          = flag.String("redis", "localhost:6379", "Redis服务器地址")
	redisPassword      = flag.String("redis-password", "", "Redis服务器密码")
	redisDB            = flag.Int("redis-db", 0, "Redis数据库编号")
	showHelp           = flag.Bool("help", false, "显示帮助信息")
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

	// 初始化缓存管理器
	if serverConfig.EnableCache {
		log.Println("初始化缓存管理器...")
		cacheManager = NewCacheManager(serverConfig.RedisAddr, serverConfig.RedisPassword, serverConfig.RedisDB)
	} else {
		cacheManager = NewCacheManager("", "", 0) // 禁用缓存
	}

	// 初始化数据库
	log.Println("初始化数据库...")
	db, err := NewDatabase(serverConfig.DatabasePath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	data.database = db

	// 加载现有服务器数据到内存缓存
	if err := loadServersFromDatabase(); err != nil {
		log.Printf("从数据库加载服务器数据失败: %v", err)
	}

	log.Println("启动 ServerStatus Monitor Data Server...")
	log.Printf("端口: %s", serverConfig.Port)
	log.Printf("数据限制: %d 条记录", serverConfig.DataLimit)
	log.Printf("推荐数据间隔: %d 秒", serverConfig.DataInterval)
	log.Printf("数据库路径: %s", serverConfig.DatabasePath)
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

	r := mux.NewRouter()

	// 添加CORS中间件支持前后端分离
	r.Use(corsMiddleware)

	// 定义需要压缩的API路径
	compressiblePaths := []string{
		"/api/servers",
		"/api/server/",
		"/api/access/",
		"/api/uuid-count",
		"/api/user-resources/",
	}

	// API路由 - 应用压缩
	api := r.PathPrefix("/api").Subrouter()
	api.Use(compressionConditionalMiddleware(compressiblePaths))

	api.HandleFunc("/data", handleData).Methods("POST")
	api.HandleFunc("/register-session", handleRegisterSession).Methods("POST")
	api.HandleFunc("/servers", handleGetServers).Methods("GET")
	api.HandleFunc("/server/{hostname}", handleGetServer).Methods("GET")
	// 双密钥认证相关路由
	api.HandleFunc("/generate-access-key", handleGenerateAccessKey).Methods("POST")
	api.HandleFunc("/access/{accessKey}/servers", handleGetServersByAccessKey).Methods("GET")
	api.HandleFunc("/access/{accessKey}/server/{hostname}", handleGetServerByAccessKey).Methods("GET")
	api.HandleFunc("/access/{accessKey}/server-by-session/{sessionID}", handleGetServerBySessionID).Methods("GET")
	api.HandleFunc("/access/{accessKey}/user-resources/{hostname}", handleGetUserResourcesByAccessKey).Methods("GET")
	api.HandleFunc("/user-resources/{hostname}", handleGetUserResources).Methods("GET")
	api.HandleFunc("/uuid-count", handleGetUUIDCount).Methods("GET")

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
	// 这里可以预加载活跃的服务器数据到内存
	// 为了性能考虑，我们只加载最新的服务器信息，历史数据按需从数据库读取
	log.Println("从数据库预加载服务器数据...")
	return nil
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

	// 使用sessionID作为key，如果没有sessionID则使用hostname（向后兼容）
	sessionKey := info.SessionID
	if sessionKey == "" {
		sessionKey = info.Hostname
	}

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
			Hostname:          info.Hostname,
			SessionID:         info.SessionID,
			LastSeen:          time.Now(),
			Status:            "online",
			CPUPercent:        info.CPU.UsagePercent,
			MemoryPercent:     info.Memory.UsagePercent,
			DiskPercent:       info.Disk.UsagePercent,
			OS:                info.OS.Platform,
			CPUTemp:           info.Temperature.CPUTemp,
			GPUTemp:           info.Temperature.GPUTemp,
			GPUs:              info.GPUs,
			MaxTemp:           info.Temperature.MaxTemp,
			NetworkSpeedSent:  info.Network.SpeedSent,
			NetworkSpeedRecv:  info.Network.SpeedRecv,
			NetworkBytesSent:  info.Network.BytesSent,
			NetworkBytesRecv:  info.Network.BytesRecv,
			UserResources:     info.UserResources,
		}

		// 广播更新到所有WebSocket客户端
		webSocketManager.BroadcastServerUpdate(serverStatus, action)
	}
}

func handleGetServers(w http.ResponseWriter, r *http.Request) {
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
				servers = getServersFromMemory(projectKey, now)
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
						Hostname:          server.Latest.Hostname,
						SessionID:         server.Latest.SessionID,
						LastSeen:          server.LastSeen,
						Status:            status,
						CPUPercent:        server.Latest.CPU.UsagePercent,
						MemoryPercent:     server.Latest.Memory.UsagePercent,
						DiskPercent:       server.Latest.Disk.UsagePercent,
						OS:                server.Latest.OS.Platform,
						CPUTemp:           server.Latest.Temperature.CPUTemp,
						GPUTemp:           server.Latest.Temperature.GPUTemp,
						GPUs:              server.Latest.GPUs,
						MaxTemp:           server.Latest.Temperature.MaxTemp,
						NetworkSpeedSent:  server.Latest.Network.SpeedSent,
						NetworkSpeedRecv:  server.Latest.Network.SpeedRecv,
						NetworkBytesSent:  server.Latest.Network.BytesSent,
						NetworkBytesRecv:  server.Latest.Network.BytesRecv,
						UserResources:     server.Latest.UserResources,
					})
			}

				// 获取总数用于分页
				totalCount, _ = data.database.GetServerCount(projectKey)

				// 缓存查询结果（仅缓存第一页）
				if useCache && serverConfig.EnableCache && page == 1 {
					go func() {
						cacheCtx := context.Background()
						cacheManager.SetServersList(cacheCtx, projectKey, servers)
						cacheManager.SetServerCount(cacheCtx, projectKey, totalCount)
						log.Printf("服务器列表已缓存: %d 个服务器", len(servers))
					}()
				}
			}
		} else {
			// 从内存获取数据
			servers = getServersFromMemory(projectKey, now)
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
func getServersFromMemory(projectKey string, now time.Time) []ServerStatus {
	data.mu.RLock()
	defer data.mu.RUnlock()

	var servers []ServerStatus

	for _, server := range data.servers {
		if server.Latest == nil {
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
			Hostname:          server.Latest.Hostname,
			SessionID:         server.Latest.SessionID,
			LastSeen:          server.LastSeen,
			Status:            status,
			CPUPercent:        server.Latest.CPU.UsagePercent,
			MemoryPercent:     server.Latest.Memory.UsagePercent,
			DiskPercent:       server.Latest.Disk.UsagePercent,
			OS:                server.Latest.OS.Platform,
			CPUTemp:           server.Latest.Temperature.CPUTemp,
			GPUTemp:           server.Latest.Temperature.GPUTemp,
			GPUs:              server.Latest.GPUs,
			MaxTemp:           server.Latest.Temperature.MaxTemp,
			NetworkSpeedSent:  server.Latest.Network.SpeedSent,
			NetworkSpeedRecv:  server.Latest.Network.SpeedRecv,
			NetworkBytesSent:  server.Latest.Network.BytesSent,
			NetworkBytesRecv:  server.Latest.Network.BytesRecv,
			UserResources:     server.Latest.UserResources,
		})
	}

	return servers
}

func handleGetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	data.mu.RLock()
	defer data.mu.RUnlock()

	server, exists := data.servers[hostname]
	if !exists {
		http.Error(w, "服务器不存在", http.StatusNotFound)
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

// handleRegisterSession 注册新的session
func handleRegisterSession(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Session注册] 收到注册请求，来源IP: %s", r.RemoteAddr)

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

	response := SessionRegisterResponse{
		SessionID: sessionID,
		Hostname:  req.Hostname,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	log.Printf("[Session注册] 成功为主机 %s 分配Session ID: %s", req.Hostname, sessionID)
}

// handleGenerateAccessKey 生成访问密钥
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

// handleGetServersByAccessKey 根据访问密钥获取服务器列表
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
			Hostname:          server.Latest.Hostname,
			SessionID:         server.Latest.SessionID,
			LastSeen:          server.LastSeen,
			Status:            status,
			CPUPercent:        server.Latest.CPU.UsagePercent,
			MemoryPercent:     server.Latest.Memory.UsagePercent,
			DiskPercent:       server.Latest.Disk.UsagePercent,
			OS:                server.Latest.OS.Platform,
			CPUTemp:           server.Latest.Temperature.CPUTemp,
			GPUTemp:           server.Latest.Temperature.GPUTemp,
			GPUs:              server.Latest.GPUs, // 添加所有GPU信息
			MaxTemp:           server.Latest.Temperature.MaxTemp,
			NetworkSpeedSent:  server.Latest.Network.SpeedSent,
			NetworkSpeedRecv:  server.Latest.Network.SpeedRecv,
			NetworkBytesSent:  server.Latest.Network.BytesSent,
			NetworkBytesRecv:  server.Latest.Network.BytesRecv,
			UserResources:     server.Latest.UserResources,
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

// handleGetServerByAccessKey 根据访问密钥和主机名获取特定服务器详情
func handleGetServerByAccessKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accessKey := vars["accessKey"]
	hostname := vars["hostname"]

	// 验证访问密钥格式
	if accessKey == "" {
		http.Error(w, "无效的访问密钥", http.StatusUnauthorized)
		return
	}

	data.mu.RLock()
	defer data.mu.RUnlock()

	// 查找匹配hostname的服务器（可能有多个session）
	var matchedServer *ServerInfo
	for _, server := range data.servers {
		if server.Latest != nil && server.Latest.Hostname == hostname {
			if isServerMatchingAccessKey(server.Latest.ProjectKey, accessKey) {
				matchedServer = server
				break
			}
		}
	}

	if matchedServer == nil {
		http.Error(w, "服务器不存在或访问被拒绝", http.StatusNotFound)
		return
	}

	// 过滤历史数据，只返回匹配访问密钥的数据
	filteredServer := &ServerInfo{
		History:  make([]*SystemInfo, 0),
		Latest:   matchedServer.Latest,
		LastSeen: matchedServer.LastSeen,
	}

	for _, historyItem := range matchedServer.History {
		if isServerMatchingAccessKey(historyItem.ProjectKey, accessKey) {
			filteredServer.History = append(filteredServer.History, historyItem)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredServer)
}

// handleGetServerBySessionID 根据访问密钥和sessionID获取特定服务器详情
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
	json.NewEncoder(w).Encode(filteredServer)
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

// handleGetUserResources 获取特定服务器的用户资源使用情况
func handleGetUserResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	data.mu.RLock()
	defer data.mu.RUnlock()

	// 查找服务器
	server, exists := data.servers[hostname]
	if !exists {
		http.Error(w, "服务器不存在", http.StatusNotFound)
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

// handleGetUserResourcesByAccessKey 根据访问密钥获取用户资源数据
func handleGetUserResourcesByAccessKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accessKey := vars["accessKey"]
	hostname := vars["hostname"]

	// 验证访问密钥格式
	if accessKey == "" {
		http.Error(w, "无效的访问密钥", http.StatusUnauthorized)
		return
	}

	data.mu.RLock()
	defer data.mu.RUnlock()

	// 查找服务器
	var matchedServer *ServerInfo
	for _, server := range data.servers {
		if server.Latest != nil && server.Latest.Hostname == hostname {
			if isServerMatchingAccessKey(server.Latest.ProjectKey, accessKey) {
				matchedServer = server
				break
			}
		}
	}

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

// handleGetUUIDCount 获取UUID数量统计（带缓存）
func handleGetUUIDCount(w http.ResponseWriter, r *http.Request) {
	var response map[string]interface{}
	var err error
	ctx := r.Context()

	// 尝试从Redis缓存获取数据
	if serverConfig.EnableCache {
		cachedResponse, err := cacheManager.GetUUIDStats(ctx)
		if err == nil && len(cachedResponse) > 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cachedResponse)
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
		json.NewEncoder(w).Encode(data.uuidStatsCache)
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
			cacheManager.SetUUIDStats(cacheCtx, response)
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
	// 读取API文档
	docPath := "./API.md"
	doc, err := os.ReadFile(docPath)
	if err != nil {
		http.Error(w, "API文档不存在", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if _, err := w.Write(doc); err != nil {
		log.Printf("Error writing API doc: %v", err)
	}

	log.Printf("API文档访问请求来自: %s", r.RemoteAddr)
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

// handleWebSocketStats 获取WebSocket连接统计信息
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

// handleCacheStats 获取缓存统计信息
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
						Hostname:          server.Latest.Hostname,
						SessionID:         server.Latest.SessionID,
						LastSeen:          server.LastSeen,
						Status:            "offline",
						CPUPercent:        server.Latest.CPU.UsagePercent,
						MemoryPercent:     server.Latest.Memory.UsagePercent,
						DiskPercent:       server.Latest.Disk.UsagePercent,
						OS:                server.Latest.OS.Platform,
						CPUTemp:           server.Latest.Temperature.CPUTemp,
						GPUTemp:           server.Latest.Temperature.GPUTemp,
						GPUs:              server.Latest.GPUs,
						MaxTemp:           server.Latest.Temperature.MaxTemp,
						NetworkSpeedSent:  server.Latest.Network.SpeedSent,
						NetworkSpeedRecv:  server.Latest.Network.SpeedRecv,
						NetworkBytesSent:  server.Latest.Network.BytesSent,
						NetworkBytesRecv:  server.Latest.Network.BytesRecv,
						UserResources:     server.Latest.UserResources,
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
		log.Printf("解析服务器配置文件失败: %v", err)
		return
	}

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
	serverConfig.RequireAuth = fileConfig.RequireAuth
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
	fmt.Println("  GET  /api/server/{hostname} - 获取特定服务器详情")
	// 已移除项目密钥和访问令牌相关API端点
	fmt.Println("  POST /api/generate-access-key - 生成访问密钥 (双密钥认证)")
	fmt.Println("  GET  /api/access/{accessKey}/servers - 根据访问密钥获取服务器列表")
	fmt.Println("  GET  /api/access/{accessKey}/server/{hostname} - 根据访问密钥获取特定服务器")
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
