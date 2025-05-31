package main

import (
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type SystemInfo struct {
	Hostname    string    `json:"hostname"`
	SessionID   string    `json:"session_id,omitempty"` // UUID session标识
	Timestamp   time.Time `json:"timestamp"`
	CPU         CPUInfo   `json:"cpu"`
	Memory      MemInfo   `json:"memory"`
	Disk        DiskInfo  `json:"disk"`
	Network     NetInfo   `json:"network"`
	GPU         GPUInfo   `json:"gpu"`  // 保持兼容性，主GPU信息
	GPUs        []GPUInfo `json:"gpus"` // 所有GPU信息
	OS          OSInfo    `json:"os"`
	Temperature TempInfo  `json:"temperature"`
	ProjectKey  string    `json:"project_key,omitempty"`
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
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
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

type ServerData struct {
	mu             sync.RWMutex
	servers        map[string]*ServerInfo // key: sessionID, value: ServerInfo
	uuidStatsCache map[string]interface{} // UUID统计缓存
	uuidCacheTime  time.Time              // 缓存更新时间
	uuidCacheMutex sync.RWMutex           // 缓存读写锁
}

type ServerInfo struct {
	Latest   *SystemInfo   `json:"latest"`
	History  []*SystemInfo `json:"history"`
	LastSeen time.Time     `json:"last_seen"`
}

type ServerStatus struct {
	Hostname      string    `json:"hostname"`
	SessionID     string    `json:"session_id,omitempty"` // UUID session标识
	LastSeen      time.Time `json:"last_seen"`
	Status        string    `json:"status"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskPercent   float64   `json:"disk_percent"`
	OS            string    `json:"os"`
	CPUTemp       float64   `json:"cpu_temp"`
	GPUTemp       float64   `json:"gpu_temp"` // 保持兼容性，主GPU温度
	GPUs          []GPUInfo `json:"gpus"`     // 所有GPU信息
	MaxTemp       float64   `json:"max_temp"`
}

type ServerConfig struct {
	ProjectKey   string `json:"project_key"`
	ServerKey    string `json:"server_key"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	RequireAuth  bool   `json:"require_auth"`
	DataLimit    int    `json:"data_limit"`    // 数据保留条数限制
	DataInterval int    `json:"data_interval"` // 数据上报间隔(秒)
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
	}

	serverConfig = ServerConfig{
		ProjectKey:   "public",           // 默认项目密钥
		ServerKey:    "serverstatus.ltd", // 默认服务器密钥
		Host:         "0.0.0.0",
		Port:         "8080",
		RequireAuth:  false,
		DataLimit:    1000, // 默认保留1000条数据
		DataInterval: 5,    // 默认5秒间隔
	}

	// 全局AccessKey缓存
	accessKeyCache = &AccessKeyCache{
		cache: make(map[string]string),
	}

	// 命令行参数
	projectKey   = flag.String("key", "", "项目认证密钥")
	serverKey    = flag.String("server-key", "", "服务器密钥 (用于双密钥认证)")
	host         = flag.String("host", "0.0.0.0", "服务器绑定IP地址")
	port         = flag.String("port", "8080", "服务器端口")
	configFile   = flag.String("config", "server-config.json", "服务器配置文件路径")
	requireAuth  = flag.Bool("auth", false, "是否要求API密钥认证")
	dataLimit    = flag.Int("data-limit", 1000, "数据保留条数限制")
	dataInterval = flag.Int("data-interval", 5, "推荐的数据上报间隔(秒)")
	showHelp     = flag.Bool("help", false, "显示帮助信息")
)

//go:embed web-ui
var webUI embed.FS

const (
	offlineThreshold = 30 * time.Second
)

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

	log.Println("启动 ServerStatus Monitor Data Server...")
	log.Printf("端口: %s", serverConfig.Port)
	log.Printf("数据限制: %d 条记录", serverConfig.DataLimit)
	log.Printf("推荐数据间隔: %d 秒", serverConfig.DataInterval)
	if serverConfig.RequireAuth {
		log.Println("API认证: 启用 (双密钥认证模式)")
	} else {
		log.Println("API认证: 禁用")
	}

	r := mux.NewRouter()

	// API路由
	r.HandleFunc("/api/data", handleData).Methods("POST")
	r.HandleFunc("/api/register-session", handleRegisterSession).Methods("POST")
	r.HandleFunc("/api/servers", handleGetServers).Methods("GET")
	r.HandleFunc("/api/server/{hostname}", handleGetServer).Methods("GET")
	// 移除基于项目密钥和访问令牌的路由，只保留AccessKey访问方式
	// 双密钥认证相关路由
	r.HandleFunc("/api/generate-access-key", handleGenerateAccessKey).Methods("POST")
	r.HandleFunc("/api/access/{accessKey}/servers", handleGetServersByAccessKey).Methods("GET")
	r.HandleFunc("/api/access/{accessKey}/server/{hostname}", handleGetServerByAccessKey).Methods("GET")
	r.HandleFunc("/api/access/{accessKey}/server-by-session/{sessionID}", handleGetServerBySessionID).Methods("GET")
	r.HandleFunc("/api/uuid-count", handleGetUUIDCount).Methods("GET")

	// 下载路由
	r.HandleFunc("/download/{filename}", handleDownload).Methods("GET")
	r.HandleFunc("/install", handleInstallScript).Methods("GET")

	// 静态文件服务 - 使用嵌入的文件系统
	webUIFS, err := fs.Sub(webUI, "web-ui")
	if err != nil {
		log.Fatal("无法创建嵌入文件系统:", err)
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(webUIFS)))

	// 启动清理协程
	go cleanupRoutine()

	log.Printf("服务器启动在 %s:%s", serverConfig.Host, serverConfig.Port)
	if serverConfig.Host == "0.0.0.0" {
		log.Printf("访问 http://localhost:%s 查看监控界面", serverConfig.Port)
		log.Printf("本地访问 http://localhost:%s 查看监控界面", serverConfig.Port)
	} else {
		log.Printf("访问 http://%s:%s 查看监控界面", serverConfig.Host, serverConfig.Port)
	}
	log.Fatal(http.ListenAndServe(serverConfig.Host+":"+serverConfig.Port, r))
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

	data.mu.Lock()
	defer data.mu.Unlock()

	// 使用sessionID作为key，如果没有sessionID则使用hostname（向后兼容）
	serverKey := info.SessionID
	if serverKey == "" {
		serverKey = info.Hostname
	}

	if data.servers[serverKey] == nil {
		data.servers[serverKey] = &ServerInfo{
			History: make([]*SystemInfo, 0, serverConfig.DataLimit),
		}
		log.Printf("新服务器注册: %s (Session: %s)", info.Hostname, serverKey)
	}

	server := data.servers[serverKey]
	server.Latest = &info
	server.LastSeen = time.Now()

	// 添加到历史记录
	server.History = append(server.History, &info)
	if len(server.History) > serverConfig.DataLimit {
		server.History = server.History[1:]
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("收到 %s 的数据上报 (Session: %s)", info.Hostname, serverKey)
}

func handleGetServers(w http.ResponseWriter, r *http.Request) {
	data.mu.RLock()
	defer data.mu.RUnlock()

	var servers []ServerStatus
	now := time.Now()

	for _, server := range data.servers {
		if server.Latest == nil {
			continue
		}

		// 默认只显示ProjectKey为"public"的服务器
		if server.Latest.ProjectKey != "public" {
			continue
		}

		status := "online"
		if now.Sub(server.Latest.Timestamp) > offlineThreshold {
			status = "offline"
		}

		servers = append(servers, ServerStatus{
			Hostname:      server.Latest.Hostname,
			SessionID:     server.Latest.SessionID,
			LastSeen:      server.LastSeen,
			Status:        status,
			CPUPercent:    server.Latest.CPU.UsagePercent,
			MemoryPercent: server.Latest.Memory.UsagePercent,
			DiskPercent:   server.Latest.Disk.UsagePercent,
			OS:            server.Latest.OS.Platform,
			CPUTemp:       server.Latest.Temperature.CPUTemp,
			GPUTemp:       server.Latest.Temperature.GPUTemp,
			GPUs:          server.Latest.GPUs, // 添加所有GPU信息
			MaxTemp:       server.Latest.Temperature.MaxTemp,
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
			Hostname:      server.Latest.Hostname,
			SessionID:     server.Latest.SessionID,
			LastSeen:      server.LastSeen,
			Status:        status,
			CPUPercent:    server.Latest.CPU.UsagePercent,
			MemoryPercent: server.Latest.Memory.UsagePercent,
			DiskPercent:   server.Latest.Disk.UsagePercent,
			OS:            server.Latest.OS.Platform,
			CPUTemp:       server.Latest.Temperature.CPUTemp,
			GPUTemp:       server.Latest.Temperature.GPUTemp,
			GPUs:          server.Latest.GPUs, // 添加所有GPU信息
			MaxTemp:       server.Latest.Temperature.MaxTemp,
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

// handleGetUUIDCount 获取UUID数量统计（带缓存）
func handleGetUUIDCount(w http.ResponseWriter, r *http.Request) {
	// 检查缓存是否有效（1分钟内）
	data.uuidCacheMutex.RLock()
	cacheValid := time.Since(data.uuidCacheTime) < time.Minute
	if cacheValid && len(data.uuidStatsCache) > 0 {
		// 使用缓存数据
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data.uuidStatsCache)
		data.uuidCacheMutex.RUnlock()
		return
	}
	data.uuidCacheMutex.RUnlock()

	// 缓存无效，重新计算
	data.mu.RLock()
	// 统计活跃的UUID数量（有session ID的服务器）
	activeUUIDs := 0
	totalServers := len(data.servers)

	for _, server := range data.servers {
		// 检查是否有有效的session ID（不是hostname）
		if server.Latest != nil && server.Latest.SessionID != "" && server.Latest.SessionID != server.Latest.Hostname {
			activeUUIDs++
		}
	}
	data.mu.RUnlock()

	// 构造响应
	response := map[string]interface{}{
		"total_servers": totalServers,
		"active_uuids":  activeUUIDs,
		"hostname_only": totalServers - activeUUIDs,
		"timestamp":     time.Now(),
		"description":   "使用我们服务的设备统计",
	}

	// 更新缓存
	data.uuidCacheMutex.Lock()
	data.uuidStatsCache = response
	data.uuidCacheTime = time.Now()
	data.uuidCacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
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

func cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		data.mu.Lock()
		now := time.Now()
		for hostname, server := range data.servers {
				if server.Latest != nil && now.Sub(server.Latest.Timestamp) > 10*time.Minute {
					log.Printf("清理长时间离线的服务器: %s", hostname)
					delete(data.servers, hostname)
				}
			}
			data.mu.Unlock()
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

	// 先检查缓存
	accessKeyCache.mu.RLock()
	if cachedKey, exists := accessKeyCache.cache[cacheKey]; exists {
		accessKeyCache.mu.RUnlock()
		return cachedKey
	}
	accessKeyCache.mu.RUnlock()

	// 缓存中不存在，计算新的AccessKey
	combinedKey := serverKey + ":" + projectKey
	hash := sha256.Sum256([]byte(combinedKey + "serverstatus-access-key-salt"))
	accessKey := hex.EncodeToString(hash[:])

	// 存入缓存
	accessKeyCache.mu.Lock()
	accessKeyCache.cache[cacheKey] = accessKey
	accessKeyCache.mu.Unlock()

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
}
