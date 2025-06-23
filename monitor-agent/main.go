package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
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
	Name          string  `json:"name"`
	MemoryTotal   uint64  `json:"memory_total"`
	MemoryUsed    uint64  `json:"memory_used"`
	UsagePercent  float64 `json:"usage_percent"`
	Temperature   float64 `json:"temperature"`
	DriverVersion string  `json:"driver_version"`
	CudaVersion   string  `json:"cuda_version"`
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
	Other   map[string]float64 `json:"other"`
	MaxTemp float64            `json:"max_temp"`
	AvgTemp float64            `json:"avg_temp"`
}

type Config struct {
	ServerURL      string        `json:"server_url"`
	ProjectKey     string        `json:"project_key"`
	ServerKey      string        `json:"server_key"`
	ReportInterval time.Duration `json:"report_interval"`
	Timeout        time.Duration `json:"timeout"`
}

var (
	config = Config{
		ServerURL:      "https://serverstatus.ltd/api/data",
		ProjectKey:     "public",
		ServerKey:      "serverstatus.ltd",
		ReportInterval: 5 * time.Second,
		Timeout:        10 * time.Second,
	}
	sessionID string // 全局session ID
	
	// 网络速率计算相关
	lastNetworkStats map[string]psnet.IOCountersStat
	lastStatsTime    time.Time
)

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

var (
	// 命令行参数
	serverURL  = flag.String("url", "", "服务器上报URL")
	projectKey = flag.String("key", "", "项目密钥 (Project Key)")
	serverKey  = flag.String("server-key", "", "服务器密钥 (Server Key) - 双密钥认证必需")
	configFile = flag.String("config", "config.json", "配置文件路径")
	silentMode = flag.Bool("silent", false, "静默模式 - 第一次上报成功后不再打印上报信息")
	showHelp   = flag.Bool("help", false, "显示帮助信息")

	// 静默模式状态
	firstReportSuccess = false
)

func main() {
	flag.Parse()

	if *showHelp {
		printUsage()
		return
	}

	// 加载配置文件
	loadConfig()

	// 命令行参数覆盖配置文件
	if *serverURL != "" {
		config.ServerURL = *serverURL
	}
	if *projectKey != "" {
		config.ProjectKey = *projectKey
	}
	if *serverKey != "" {
		config.ServerKey = *serverKey
	}

	log.Println("启动 ServerStatus Monitor Agent...")
	log.Println("📦 项目地址 | Project Repository: https://github.com/MyDailyCloud/ServerStatus")
	log.Println("⭐ 如果觉得有用，请给个Star支持一下 | If you find it useful, please give us a Star!")

	// 强制要求双密钥认证
	if config.ProjectKey == "" || config.ServerKey == "" {
		log.Println("❌ 错误: 双密钥认证要求同时提供主密钥和团队密钥")
		log.Println("使用方法:")
		log.Println("  monitor-agent -url <server-url> -key <project-key> -server-key <server-key>")
		log.Println("示例:")
		log.Println("  monitor-agent -url http://192.168.1.100:8080/api/data -key project-alpha -server-key your-server-secret")
		os.Exit(1)
	}

	hostname, _ := os.Hostname()
	log.Printf("主机名 | Hostname: %s", hostname)
	log.Printf("上报地址 | Report URL: %s", config.ServerURL)
	log.Printf("使用项目密钥 | Using project key: %s...", config.ProjectKey[:min(8, len(config.ProjectKey))])
	log.Printf("使用服务器密钥 | Using server key: %s...", config.ServerKey[:min(8, len(config.ServerKey))])

	// 自动生成访问链接
	generateAccessLinks()

	// 注册session获取UUID
	if err := registerSession(); err != nil {
		log.Printf("Session注册失败，将使用hostname作为标识 | Session registration failed, will use hostname as identifier: %v", err)
		sessionID = "" // 清空sessionID，使用hostname作为fallback
	}

	log.Printf("上报间隔 | Report interval: %v", config.ReportInterval)

	ticker := time.NewTicker(config.ReportInterval)
	defer ticker.Stop()

	// 立即发送一次
	collectAndReport()

	for range ticker.C {
		collectAndReport()
	}
}

func collectAndReport() {
	info, err := collectSystemInfo()
	if err != nil {
		log.Printf("收集系统信息失败 | Failed to collect system info: %v", err)
		return
	}

	err = reportToServer(info)
	if err != nil {
		log.Printf("上报数据失败 | Failed to report data: %v", err)
	} else {
		// 静默模式逻辑
		if *silentMode {
			if !firstReportSuccess {
				// 第一次成功上报，打印详细信息
				gpuInfo := "无GPU | No GPU"
				if len(info.GPUs) > 0 {
					gpuInfo = fmt.Sprintf("%d个GPU | %d GPUs: %s (%.1f°C)", len(info.GPUs), len(info.GPUs), info.GPUs[0].Name, info.GPUs[0].Temperature)
				}
				log.Printf("首次上报成功 | First report successful - CPU: %.1f%%, 内存 | Memory: %.1f%%, 磁盘 | Disk: %.1f%%, GPU: %s",
					info.CPU.UsagePercent, info.Memory.UsagePercent, info.Disk.UsagePercent, gpuInfo)
				log.Println("静默模式已启用，后续上报将不再显示详细信息 | Silent mode enabled, subsequent reports will not show details")
				firstReportSuccess = true
			}
			// 静默模式下后续上报不打印任何信息
		} else {
			// 非静默模式，正常打印详细信息
			gpuInfo := "无GPU | No GPU"
			if len(info.GPUs) > 0 {
				gpuInfo = fmt.Sprintf("%d个GPU | %d GPUs: %s (%.1f°C)", len(info.GPUs), len(info.GPUs), info.GPUs[0].Name, info.GPUs[0].Temperature)
			}
			log.Printf("成功上报数据 | Data reported successfully - CPU: %.1f%%, 内存 | Memory: %.1f%%, 磁盘 | Disk: %.1f%%, GPU: %s",
				info.CPU.UsagePercent, info.Memory.UsagePercent, info.Disk.UsagePercent, gpuInfo)
		}
	}
}

// registerSession 注册session获取UUID
func registerSession() error {
	hostname, _ := os.Hostname()

	req := SessionRegisterRequest{
		Hostname:   hostname,
		ProjectKey: config.ProjectKey,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("编码注册请求失败 | Failed to encode register request: %v", err)
	}

	// 构造注册URL，将/api/data替换为/api/register-session
	baseURL := strings.Replace(config.ServerURL, "/api/data", "", 1)
	url := baseURL + "/api/register-session"
	client := &http.Client{
		Timeout: config.Timeout,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("注册session失败 | Failed to register session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("注册session失败，状态码 | Failed to register session, status code: %d", resp.StatusCode)
	}

	var response SessionRegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析注册响应失败 | Failed to decode register response: %v", err)
	}

	sessionID = response.SessionID
	log.Printf("Session注册成功 | Session registered successfully: %s", sessionID)
	return nil
}

func collectSystemInfo() (*SystemInfo, error) {
	hostname, _ := os.Hostname()

	info := &SystemInfo{
		Hostname:   hostname,
		SessionID:  sessionID,
		Timestamp:  time.Now(),
		ProjectKey: config.ProjectKey,
	}

	// CPU信息
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		info.CPU.UsagePercent = cpuPercent[0]
	}
	info.CPU.CoreCount = runtime.NumCPU()

	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		info.CPU.ModelName = cpuInfos[0].ModelName
	}

	// 内存信息
	memStat, err := mem.VirtualMemory()
	if err == nil {
		info.Memory.Total = memStat.Total
		info.Memory.Used = memStat.Used
		info.Memory.Free = memStat.Free
		info.Memory.UsagePercent = memStat.UsedPercent
	}

	// 磁盘信息
	diskStat, err := disk.Usage("/")
	if runtime.GOOS == "windows" {
		diskStat, err = disk.Usage("C:")
	}
	if err == nil {
		info.Disk.Total = diskStat.Total
		info.Disk.Used = diskStat.Used
		info.Disk.Free = diskStat.Free
		info.Disk.UsagePercent = diskStat.UsedPercent
	}

	// 网络信息
	info.Network = collectNetworkInfo()

	// GPU信息
	gpuInfos := collectGPUInfo()
	info.GPUs = gpuInfos // 存储所有GPU信息
	if len(gpuInfos) > 0 {
		info.GPU = gpuInfos[0] // 使用第一个GPU的信息（保持兼容性）
	} else {
		info.GPU = GPUInfo{
			Name:         "Unknown GPU",
			MemoryTotal:  0,
			MemoryUsed:   0,
			UsagePercent: 0,
			Temperature:  0,
		}
	}

	// 操作系统信息
	hostInfo, err := host.Info()
	if err == nil {
		info.OS.Platform = hostInfo.Platform
		info.OS.Version = hostInfo.PlatformVersion
		info.OS.Arch = hostInfo.KernelArch
		info.OS.Uptime = hostInfo.Uptime
	}

	// 温度信息
	info.Temperature = collectTemperatureInfo()

	return info, nil
}

func reportToServer(info *SystemInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	req, err := http.NewRequest("POST", config.ServerURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if config.ProjectKey != "" {
		req.Header.Set("X-Project-Key", config.ProjectKey)
	}
	if config.ServerKey != "" {
		req.Header.Set("X-Server-Key", config.ServerKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	return nil
}

// collectTemperatureInfo 收集温度信息
func collectTemperatureInfo() TempInfo {
	tempInfo := TempInfo{
		Other: make(map[string]float64),
	}

	// 在Linux系统上尝试使用sensors命令
	if runtime.GOOS == "linux" {
		if temps := getLinuxTemperatures(); len(temps) > 0 {
			var cpuTemps []float64
			var gpuTemps []float64
			var allTemps []float64

			for name, temp := range temps {
				allTemps = append(allTemps, temp)
				tempInfo.Other[name] = temp

				// 判断是否为CPU温度
				if strings.Contains(strings.ToLower(name), "cpu") ||
					strings.Contains(strings.ToLower(name), "core") ||
					strings.Contains(strings.ToLower(name), "package") {
					cpuTemps = append(cpuTemps, temp)
				}

				// 判断是否为GPU温度
				if strings.Contains(strings.ToLower(name), "gpu") ||
					strings.Contains(strings.ToLower(name), "nvidia") ||
					strings.Contains(strings.ToLower(name), "radeon") {
					gpuTemps = append(gpuTemps, temp)
				}
			}

			// 计算平均温度
			if len(cpuTemps) > 0 {
				var sum float64
				for _, temp := range cpuTemps {
					sum += temp
					if temp > tempInfo.CPUTemp {
						tempInfo.CPUTemp = temp
					}
				}
				tempInfo.CPUTemp = sum / float64(len(cpuTemps))
			}

			if len(gpuTemps) > 0 {
				var sum float64
				for _, temp := range gpuTemps {
					sum += temp
					if temp > tempInfo.GPUTemp {
						tempInfo.GPUTemp = temp
					}
				}
				tempInfo.GPUTemp = sum / float64(len(gpuTemps))
			}

			// 计算最高温度和平均温度
			if len(allTemps) > 0 {
				var sum float64
				for _, temp := range allTemps {
					sum += temp
					if temp > tempInfo.MaxTemp {
						tempInfo.MaxTemp = temp
					}
				}
				tempInfo.AvgTemp = sum / float64(len(allTemps))
			}
		}
	}

	return tempInfo
}

// getLinuxTemperatures 在Linux系统上获取温度信息
func getLinuxTemperatures() map[string]float64 {
	temps := make(map[string]float64)

	// 尝试使用sensors命令
	cmd := exec.Command("sensors", "-A", "-u")
	output, err := cmd.Output()
	if err != nil {
		// 如果sensors命令不可用，尝试读取/sys/class/thermal
		return getThermalZoneTemperatures()
	}

	lines := strings.Split(string(output), "\n")
	var currentChip string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检测芯片名称
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "_") {
			currentChip = strings.TrimSuffix(line, ":")
			continue
		}

		// 解析温度值
		if strings.Contains(line, "_input:") && strings.Contains(line, "temp") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if temp, err := strconv.ParseFloat(parts[1], 64); err == nil {
					// sensors输出的温度单位是摄氏度，但是是毫摄氏度
					temp = temp / 1000.0
					if temp > 0 && temp < 150 { // 合理的温度范围
						sensorName := fmt.Sprintf("%s_%s", currentChip, strings.Split(parts[0], "_")[0])
						temps[sensorName] = temp
					}
				}
			}
		}
	}

	return temps
}

// getThermalZoneTemperatures 从/sys/class/thermal读取温度
func getThermalZoneTemperatures() map[string]float64 {
	temps := make(map[string]float64)

	// 尝试读取thermal zone
	for i := 0; i < 10; i++ {
		tempPath := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/temp", i)
		typePath := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/type", i)

		if tempData, err := os.ReadFile(tempPath); err == nil {
			if tempStr := strings.TrimSpace(string(tempData)); tempStr != "" {
				if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
					// thermal zone的温度单位是毫摄氏度
					temp = temp / 1000.0
					if temp > 0 && temp < 150 {
						zoneName := fmt.Sprintf("thermal_zone%d", i)
						if typeData, err := os.ReadFile(typePath); err == nil {
							if zoneType := strings.TrimSpace(string(typeData)); zoneType != "" {
								zoneName = zoneType
							}
						}
						temps[zoneName] = temp
					}
				}
			}
		}
	}

	return temps
}

// collectGPUInfo 收集GPU信息
func collectGPUInfo() []GPUInfo {
	var gpuInfos []GPUInfo

	// 获取CUDA和驱动版本信息
	driverVersion := ""
	cudaVersion := ""

	// 获取驱动版本
	if driverCmd := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader,nounits"); driverCmd != nil {
		if output, err := driverCmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				driverVersion = strings.TrimSpace(lines[0])
			}
		}
	}

	// 获取CUDA版本
	if cudaCmd := exec.Command("nvidia-smi", "--query-gpu=cuda_version", "--format=csv,noheader,nounits"); cudaCmd != nil {
		if output, err := cudaCmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				cudaVersion = strings.TrimSpace(lines[0])
			}
		}
	}

	// 尝试使用nvidia-smi命令 (支持Windows和Linux)
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,memory.used,temperature.gpu,utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Split(line, ", ")
			if len(parts) >= 5 {
				var gpuInfo GPUInfo
				gpuInfo.Name = strings.TrimSpace(parts[0])
				gpuInfo.DriverVersion = driverVersion
				gpuInfo.CudaVersion = cudaVersion

				if total, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64); err == nil {
					gpuInfo.MemoryTotal = total * 1024 * 1024 // MB to bytes
				}

				if used, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64); err == nil {
					gpuInfo.MemoryUsed = used * 1024 * 1024 // MB to bytes
				}

				if temp, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64); err == nil {
					gpuInfo.Temperature = temp
				}

				// 正确获取GPU使用率（第5个字段是utilization.gpu）
				if usage, err := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64); err == nil {
					gpuInfo.UsagePercent = usage
				}

				gpuInfos = append(gpuInfos, gpuInfo)
			}
		}
	} else {
		// 如果nvidia-smi不可用，尝试其他方法
		if runtime.GOOS == "windows" {
			// 在Windows上尝试使用WMI查询
			if gpuInfo := getWindowsGPUInfo(); gpuInfo != nil {
				gpuInfos = append(gpuInfos, *gpuInfo)
			}
		}
	}

	return gpuInfos
}

// getWindowsGPUInfo 在Windows上获取GPU信息
func getWindowsGPUInfo() *GPUInfo {
	// 尝试使用PowerShell查询GPU信息
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_VideoController | Select-Object Name, AdapterRAM | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// 解析JSON输出
	var gpuData interface{}
	err = json.Unmarshal(output, &gpuData)
	if err != nil {
		return nil
	}

	// 处理单个GPU或多个GPU的情况
	switch data := gpuData.(type) {
	case map[string]interface{}:
		// 单个GPU
		if name, ok := data["Name"].(string); ok {
			gpuInfo := &GPUInfo{
				Name: name,
			}
			if ram, ok := data["AdapterRAM"].(float64); ok && ram > 0 {
				gpuInfo.MemoryTotal = uint64(ram)
			}
			return gpuInfo
		}
	case []interface{}:
		// 多个GPU，返回第一个
		if len(data) > 0 {
			if gpu, ok := data[0].(map[string]interface{}); ok {
				if name, ok := gpu["Name"].(string); ok {
					gpuInfo := &GPUInfo{
						Name: name,
					}
					if ram, ok := gpu["AdapterRAM"].(float64); ok && ram > 0 {
						gpuInfo.MemoryTotal = uint64(ram)
					}
					return gpuInfo
				}
			}
		}
	}

	return nil
}

// loadConfig 加载配置文件
func loadConfig() {
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置文件
		saveConfig()
		log.Printf("创建默认配置文件 | Creating default config file: %s", *configFile)
		return
	}

	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Printf("读取配置文件失败 | Failed to read config file: %v", err)
		return
	}

	var fileConfig Config
	err = json.Unmarshal(data, &fileConfig)
	if err != nil {
		log.Printf("解析配置文件失败 | Failed to parse config file: %v", err)
		return
	}

	// 更新配置
	if fileConfig.ServerURL != "" {
		config.ServerURL = fileConfig.ServerURL
	}
	if fileConfig.ProjectKey != "" {
		config.ProjectKey = fileConfig.ProjectKey
	}
	if fileConfig.ReportInterval > 0 {
		config.ReportInterval = fileConfig.ReportInterval
	}
	if fileConfig.Timeout > 0 {
		config.Timeout = fileConfig.Timeout
	}

	log.Printf("加载配置文件 | Loading config file: %s", *configFile)

	// 保存配置到文件
	saveConfig()
}

func saveConfig() {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("序列化配置失败 | Failed to serialize config: %v", err)
		return
	}

	// 确保配置目录存在
	dir := filepath.Dir(*configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("创建配置目录失败 | Failed to create config directory: %v", err)
		return
	}

	err = os.WriteFile(*configFile, data, 0644)
	if err != nil {
		log.Printf("保存配置文件失败 | Failed to save config file: %v", err)
		return
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("ServerStatus Monitor Agent - 系统监控客户端 | System Monitoring Client")
	fmt.Println("📦 项目地址 | Project Repository: https://github.com/MyDailyCloud/ServerStatus")
	fmt.Println("⭐ 如果觉得有用，请给个Star支持一下 | If you find it useful, please give us a Star!")
	fmt.Println()
	fmt.Println("用法 | Usage:")
	fmt.Println("  monitor-agent [选项 | options]")
	fmt.Println()
	fmt.Println("选项 | Options:")
	fmt.Println("  -url string")
	fmt.Println("        服务器上报URL | Server report URL (例如 | e.g.: http://your-server:8080/api/data)")
	fmt.Println("  -key string")
	fmt.Println("        API认证密钥 | API authentication key")
	fmt.Println("  -server-key string")
	fmt.Println("        服务器密钥 | Server key (双密钥认证必需 | Required for dual-key authentication)")
	fmt.Println("  -config string")
	fmt.Println("        配置文件路径 | Config file path (默认 | default: config.json)")
	fmt.Println("  -silent")
	fmt.Println("        静默模式 | Silent mode - 第一次上报成功后不再打印上报信息 | Stop printing report details after first successful report")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息 | Show this help message")
	fmt.Println()
	fmt.Println("功能特性 | Features:")
	fmt.Println("  🌐 自动生成访问链接 | Auto-generate access links - 启动时自动显示监控面板访问地址 | Automatically display monitoring panel access URLs on startup")
	fmt.Println("  🔑 多种认证方式 | Multiple authentication methods - 支持API密钥、访问令牌、双密钥认证 | Support API key, access token, dual-key authentication")
	fmt.Println("  📊 实时监控 | Real-time monitoring - CPU、内存、磁盘、GPU、温度等系统信息 | CPU, memory, disk, GPU, temperature and other system information")
	fmt.Println()
	fmt.Println("示例 | Examples:")
	fmt.Println("  # 使用默认配置 | Use default config")
	fmt.Println("  monitor-agent")
	fmt.Println()
	fmt.Println("  # 双密钥认证 | Dual-key authentication (启动后自动显示访问链接 | Auto-display access links after startup)")
	fmt.Println("  monitor-agent -url http://192.168.1.100:8080/api/data -key project-alpha -server-key your-server-secret")
	fmt.Println()
	fmt.Println("环境变量设置 | Environment variable setup:")
	fmt.Println("  export SERVER_KEY=your-server-secret")
	fmt.Println("  monitor-agent -url http://192.168.1.100:8080/api/data -key project-alpha -server-key $SERVER_KEY")
	fmt.Println()
	fmt.Println("  # 使用自定义配置文件 | Use custom config file")
	fmt.Println("  monitor-agent -config /path/to/config.json")
	fmt.Println()
	fmt.Println("配置文件格式 | Config file format (JSON):")
	fmt.Println(`  {`)
	fmt.Println(`    "server_url": "http://your-server:8080/api/data",`)
	fmt.Println(`    "project_key": "project-alpha",`)
	fmt.Println(`    "server_key": "your-server-secret",`)
	fmt.Println(`    "report_interval": "1s",`)
	fmt.Println(`    "timeout": "10s"`)
	fmt.Println(`  }`)
	fmt.Println()
	fmt.Println("前后端分离架构说明 | Frontend-Backend Separation Architecture:")
	fmt.Println("  系统采用前后端分离设计，支持多种前端技术栈 | System uses frontend-backend separation, supports multiple frontend frameworks")
	fmt.Println("  • API服务器 | API Server: 提供RESTful API接口 | Provides RESTful API interfaces")
	fmt.Println("  • 前端UI | Frontend UI: 独立部署的Web界面 | Independently deployed web interface")
	fmt.Println("  • 访问方式 | Access Methods:")
	fmt.Println("    - 公开模式 | Public Mode: 无需认证，显示public项目数据")
	fmt.Println("    - 项目密钥 | Project Key: 基于项目密钥的数据隔离")
	fmt.Println("    - 访问密钥 | Access Key: 双密钥认证生成的安全访问密钥")
	fmt.Println("")
	fmt.Println("前端部署 | Frontend Deployment:")
	fmt.Println("  1. 官方UI | Official UI: cd frontend-ui && ./deploy.sh")
	fmt.Println("  2. 自定义开发 | Custom Development: 基于API开发任意前端界面")
	fmt.Println("  3. 第三方集成 | Third-party Integration: React/Vue/Angular等框架")
	fmt.Println()
	fmt.Println("安全提示 | Security tips:")
	fmt.Println("  • 主密钥应由管理员统一管理 | Master key should be managed by administrators")
	fmt.Println("  • 建议使用环境变量设置主密钥 | Recommend using environment variables for master key")
	fmt.Println("  • 不同团队使用不同的团队密钥 | Different teams should use different team keys")
}

// generateAccessLinks 生成并显示访问信息（前后端分离版本）
func generateAccessLinks() {
	// 从上报URL提取服务器地址
	serverBaseURL := extractServerBaseURL(config.ServerURL)
	if serverBaseURL == "" {
		log.Println("无法从上报URL提取服务器地址")
		return
	}

	log.Println("")
	log.Println("=== 🌐 监控访问信息 | Monitoring Access Info ===")
	
	// 显示API服务器信息
	log.Printf("📡 API服务器 | API Server: %s", serverBaseURL)
	log.Printf("📄 API文档 | API Documentation: %s/API.md", serverBaseURL)

	// 如果是公开模式
	if config.ProjectKey == "public" || config.ProjectKey == "demo" {
		log.Println("")
		log.Println("🔓 公开模式 | Public Mode:")
		log.Printf("   ✅ 项目密钥 | Project Key: %s", config.ProjectKey)
		log.Println("   📊 数据将在公开面板中显示 | Data will be shown in public panel")
		log.Println("")
		log.Println("📱 前端访问 | Frontend Access:")
		log.Println("   1. 部署前端UI | Deploy Frontend UI:")
		log.Println("      cd frontend-ui && ./deploy.sh")
		log.Println("   2. 或访问在线演示 | Or visit online demo:")
		log.Println("      https://serverstatus.ltd (if available)")
		return
	}

	// 如果启用双密钥认证
	if config.ProjectKey != "" && config.ServerKey != "" {
		log.Println("")
		log.Println("🔐 双密钥认证模式 | Dual-Key Authentication Mode:")
		log.Printf("   ✅ 服务器密钥 | Server Key: %s...", config.ServerKey[:min(8, len(config.ServerKey))])
		log.Printf("   ✅ 项目密钥 | Project Key: %s", config.ProjectKey)

		// 生成访问密钥
		if accessKey := generateAccessKey(serverBaseURL); accessKey != "" {
			log.Println("")
			log.Println("🔑 访问密钥 | Access Key:")
			log.Printf("   %s", accessKey)
			log.Println("")
			log.Println("📱 使用步骤 | Usage Steps:")
			log.Println("   1. 复制上述访问密钥 | Copy the access key above")
			log.Println("   2. 部署前端UI | Deploy Frontend UI:")
			log.Println("      cd frontend-ui && ./deploy.sh") 
			log.Println("   3. 在前端页面输入访问密钥 | Enter access key in frontend")
			log.Println("      或在URL中使用 | Or use in URL: ?key=<access-key>")
		}
	} else if config.ProjectKey != "" {
		log.Println("")
		log.Println("🔓 项目密钥模式 | Project Key Mode:")
		log.Printf("   ✅ 项目密钥 | Project Key: %s", config.ProjectKey)
		log.Println("")
		log.Println("📱 使用步骤 | Usage Steps:")
		log.Println("   1. 部署前端UI | Deploy Frontend UI:")
		log.Println("      cd frontend-ui && ./deploy.sh")
		log.Println("   2. 在前端页面输入项目密钥 | Enter project key in frontend")
		log.Println("      或在URL中使用 | Or use in URL: ?key=<project-key>")
	}

	log.Println("")
	log.Println("🛠️  API开发 | API Development:")
	log.Printf("   📖 查看API文档 | View API docs: %s/API.md", serverBaseURL)
	log.Printf("   🔌 获取服务器列表 | Get servers: %s/api/servers", serverBaseURL)
	if config.ProjectKey != "" && config.ServerKey != "" {
		log.Printf("   🔑 生成访问密钥 | Generate access key: %s/api/generate-access-key", serverBaseURL)
	}
	
	log.Println("")
	log.Println("=======================================")
	log.Println("")
}

// extractServerBaseURL 从上报URL提取服务器基础地址
func extractServerBaseURL(reportURL string) string {
	if reportURL == "" {
		return ""
	}

	// 移除 /api/data 后缀
	if strings.HasSuffix(reportURL, "/api/data") {
		return strings.TrimSuffix(reportURL, "/api/data")
	}

	// 如果URL不是标准格式，尝试提取基础部分
	parts := strings.Split(reportURL, "/api/")
	if len(parts) > 0 {
		return parts[0]
	}

	return reportURL
}

// generateTokenLink 尝试生成访问令牌
func generateTokenLink(serverBaseURL string) string {
	if config.ProjectKey == "" {
		return ""
	}

	// 构造生成令牌的请求
	tokenURL := serverBaseURL + "/api/generate-token"
	requestBody := map[string]string{
		"project_key": config.ProjectKey,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return ""
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(tokenURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return ""
	}

	return tokenResponse.Token
}

// generateAccessKey 生成访问密钥
func generateAccessKey(serverBaseURL string) string {
	accessKeyURL := serverBaseURL + "/api/generate-access-key"

	requestBody := map[string]string{
		"server_key":  config.ServerKey,
		"project_key": config.ProjectKey,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return ""
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(accessKeyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var accessKeyResponse struct {
		AccessKey string `json:"access_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&accessKeyResponse); err != nil {
		return ""
	}

	return accessKeyResponse.AccessKey
}

// collectNetworkInfo 收集详细的网络信息
func collectNetworkInfo() NetInfo {
	var netInfo NetInfo
	currentTime := time.Now()
	
	// 获取总的网络统计信息
	allStats, err := psnet.IOCounters(false)
	if err == nil && len(allStats) > 0 {
		netInfo.BytesSent = allStats[0].BytesSent
		netInfo.BytesRecv = allStats[0].BytesRecv
		netInfo.PacketsSent = allStats[0].PacketsSent
		netInfo.PacketsRecv = allStats[0].PacketsRecv
	}
	
	// 获取各个网卡的详细信息
	perInterfaceStats, err := psnet.IOCounters(true)
	if err == nil {
		// 初始化lastNetworkStats映射
		if lastNetworkStats == nil {
			lastNetworkStats = make(map[string]psnet.IOCountersStat)
		}
		
		for _, stat := range perInterfaceStats {
			// 跳过回环接口和无流量的接口
			if stat.Name == "lo" || stat.Name == "Loopback" ||
				(stat.BytesSent == 0 && stat.BytesRecv == 0) {
				continue
			}
			
			netInterface := NetInterface{
				Name:        stat.Name,
				BytesSent:   stat.BytesSent,
				BytesRecv:   stat.BytesRecv,
				PacketsSent: stat.PacketsSent,
				PacketsRecv: stat.PacketsRecv,
				IsUp:        true, // gopsutil不直接提供状态，默认为true
			}
			
			// 计算网速（如果有之前的数据）
			if lastStat, exists := lastNetworkStats[stat.Name]; exists && !lastStatsTime.IsZero() {
				timeDiff := currentTime.Sub(lastStatsTime).Seconds()
				if timeDiff > 0 {
					bytesSentDiff := stat.BytesSent - lastStat.BytesSent
					bytesRecvDiff := stat.BytesRecv - lastStat.BytesRecv
					
					// 计算速率 (KB/s)
					netInterface.SpeedSent = float64(bytesSentDiff) / timeDiff / 1024
					netInterface.SpeedRecv = float64(bytesRecvDiff) / timeDiff / 1024
					
					// 累加到总速率
					netInfo.SpeedSent += netInterface.SpeedSent
					netInfo.SpeedRecv += netInterface.SpeedRecv
				}
			}
			
			// 获取IP地址和其他接口信息
			interfaces, err := net.Interfaces()
			if err == nil {
				for _, iface := range interfaces {
					if iface.Name == stat.Name {
						netInterface.MTU = iface.MTU
						
						// 获取IP地址
						addrs, err := iface.Addrs()
						if err == nil {
							for _, addr := range addrs {
								netInterface.Addrs = append(netInterface.Addrs, addr.String())
							}
						}
						break
					}
				}
			}
			
			netInfo.Interfaces = append(netInfo.Interfaces, netInterface)
			
			// 更新缓存
			lastNetworkStats[stat.Name] = stat
		}
	}
	
	// 更新时间戳
	lastStatsTime = currentTime
	
	return netInfo
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
