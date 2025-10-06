package models

import "time"

// SystemInfo 系统信息结构
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

// CPUInfo CPU信息
type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	CoreCount    int     `json:"core_count"`
	ModelName    string  `json:"model_name"`
}

// MemInfo 内存信息
type MemInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetInfo 网络信息
type NetInfo struct {
	BytesSent    uint64        `json:"bytes_sent"`     // 总发送字节数
	BytesRecv    uint64        `json:"bytes_recv"`     // 总接收字节数
	PacketsSent  uint64        `json:"packets_sent"`   // 总发送包数
	PacketsRecv  uint64        `json:"packets_recv"`   // 总接收包数
	SpeedSent    float64       `json:"speed_sent"`     // 发送速率 (KB/s)
	SpeedRecv    float64       `json:"speed_recv"`     // 接收速率 (KB/s)
	Interfaces   []NetInterface `json:"interfaces"`     // 网卡详细信息
	TxBytes      uint64        `json:"tx_bytes"`       // 发送字节数 (兼容字段)
	RxBytes      uint64        `json:"rx_bytes"`       // 接收字节数 (兼容字段)
}

// NetInterface 网卡信息
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

// GPUInfo GPU信息
type GPUInfo struct {
	Name         string  `json:"name"`
	MemoryTotal  uint64  `json:"memory_total"`
	MemoryUsed   uint64  `json:"memory_used"`
	UsagePercent float64 `json:"usage_percent"`
	Temperature  float64 `json:"temperature"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	Architecture string `json:"architecture"`
	Hostname    string `json:"hostname"`
	Uptime      string `json:"uptime"`
	Arch        string `json:"arch"` // 兼容字段
}

// TempInfo 温度信息
type TempInfo struct {
	CPUTemp float64            `json:"cpu_temp"`
	GPUTemp float64            `json:"gpu_temp"`
	Sensors map[string]float64 `json:"sensors"`
	MaxTemp float64            `json:"max_temp"`
	AvgTemp float64            `json:"avg_temp"`
	Main    float64            `json:"main"` // 主温度 (兼容字段)
}

// UserResourceInfo 用户资源使用信息
type UserResourceInfo struct {
	Username     string     `json:"username"`
	ProcessCount int        `json:"process_count"`
	CPUUsage     float64    `json:"cpu_usage"`
	MemoryUsed   uint64     `json:"memory_used"`
	MemoryUsage  float64    `json:"memory_usage"`
	Processes    []Process  `json:"processes"`
	Timestamp    time.Time  `json:"timestamp"`
	Hostname     string     `json:"hostname"`
	SessionID    string     `json:"session_id"`
	ProjectKey   string     `json:"project_key"`
}

// Process 进程信息
type Process struct {
	Name        string  `json:"name"`
	PID         int     `json:"pid"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
}

// ServerStatus 服务器状态
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
	GPUTemp           float64             `json:"gpu_temp"`
	NetworkSpeed      float64             `json:"network_speed"`
	ProcessCount      int                 `json:"process_count"`
	SystemInfo        SystemInfo          `json:"system_info"`
	ProjectKey        string              `json:"project_key,omitempty"`
}

// ServerInfo 服务器信息 (数据库使用)
type ServerInfo struct {
	Latest    *SystemInfo `json:"latest"`
	History   []*SystemInfo `json:"history"`
	LastSeen  time.Time   `json:"last_seen"`
}

// HistoryData 历史数据
type HistoryData struct {
	Timestamp       time.Time `json:"timestamp"`
	Hostname        string    `json:"hostname"`
	SessionID       string    `json:"session_id"`
	ProjectKey      string    `json:"project_key"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsed      uint64    `json:"memory_used"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsed        uint64    `json:"disk_used"`
	DiskUsage       float64   `json:"disk_usage"`
	NetworkTx       uint64    `json:"network_tx"`
	NetworkRx       uint64    `json:"network_rx"`
	GPUUsage        float64   `json:"gpu_usage"`
	GPUMemoryUsage  float64   `json:"gpu_memory_usage"`
	GPUTemperature  float64   `json:"gpu_temperature"`
	Temperature     float64   `json:"temperature"`
	TimeBucket      string    `json:"time_bucket,omitempty"` // 用于聚合数据的时间桶
}