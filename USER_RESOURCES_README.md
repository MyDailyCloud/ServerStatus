# 用户资源监控功能说明

## 功能概述

ServerStatus 系统新增了用户资源监控功能，可以收集和展示服务器上各个用户的资源使用情况，包括：
- 每个用户的进程数量
- CPU使用率
- 内存使用量和百分比  
- TOP进程列表

## 功能特性

1. **实时监控**: 实时收集各用户的资源使用情况
2. **自动聚合**: 自动聚合每个用户的所有进程资源
3. **TOP进程**: 展示每个用户消耗资源最多的前5个进程
4. **可配置**: 可通过配置文件或命令行参数控制是否启用
5. **高性能**: 只收集和传输必要的统计数据，对系统影响最小

## 使用方法

### 1. Monitor Agent 配置

#### 命令行参数方式
```bash
# 启用用户资源监控（默认启用）
./monitor-agent -url http://server:8080/api/data -key project-key -server-key server-key

# 显式启用
./monitor-agent -url http://server:8080/api/data -key project-key -server-key server-key -user-resources

# 禁用用户资源监控
./monitor-agent -url http://server:8080/api/data -key project-key -server-key server-key -user-resources=false
```

#### 配置文件方式
创建 `config.json`:
```json
{
  "server_url": "http://localhost:8080/api/data",
  "project_key": "public",
  "server_key": "serverstatus.ltd",
  "report_interval": "5s",
  "timeout": "10s",
  "enable_user_resources": true
}
```

然后运行：
```bash
./monitor-agent -config config.json
```

### 2. Data Server

Data Server 会自动处理和存储用户资源数据，无需额外配置。

新增的API端点：
- `GET /api/user-resources/{hostname}` - 获取指定服务器的用户资源数据
- `GET /api/access/{accessKey}/user-resources/{hostname}` - 带访问密钥的用户资源访问

### 3. 前端展示

在前端UI中，当你点击某个服务器查看详情时，会自动显示用户资源使用表格，包含：
- 用户名
- 进程数量
- CPU使用率（高于50%会高亮显示）
- 内存使用量（MB和百分比）
- TOP进程列表（显示前3个最消耗资源的进程）

## 数据结构

### UserResourceInfo
```go
type UserResourceInfo struct {
    Username      string        `json:"username"`
    UID           uint32        `json:"uid"`
    ProcessCount  int           `json:"process_count"`
    CPUPercent    float64       `json:"cpu_percent"`
    MemoryMB      uint64        `json:"memory_mb"`
    MemoryPercent float64       `json:"memory_percent"`
    TopProcesses  []ProcessInfo `json:"top_processes"`
}
```

### ProcessInfo
```go
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
```

## 性能优化

1. **数据限制**: 默认只返回前20个用户（按CPU使用率排序）
2. **进程限制**: 每个用户只保留前5个TOP进程
3. **可选功能**: 可以通过配置禁用此功能以节省资源
4. **缓存机制**: Data Server会缓存用户资源数据，减少重复计算

## 兼容性

- **向后兼容**: 新功能完全向后兼容，不影响现有功能
- **平台支持**: 支持Linux、macOS、Windows等所有gopsutil支持的平台
- **权限要求**: 某些系统可能需要特定权限才能访问其他用户的进程信息

## 故障排查

1. **没有用户资源数据**
   - 检查monitor-agent是否启用了user-resources功能
   - 检查agent是否有足够权限访问进程信息
   - 查看agent日志是否有错误信息

2. **数据不完整**
   - 某些系统可能限制非root用户访问其他用户的进程信息
   - 可以尝试以更高权限运行monitor-agent

3. **性能问题**
   - 如果系统进程过多，可以考虑禁用此功能
   - 调整report_interval以减少数据收集频率

## 测试

运行测试脚本：
```bash
./test-user-resources.sh
```

测试脚本会：
1. 启动data-server
2. 启动monitor-agent并启用用户资源监控
3. 测试API端点
4. 验证用户资源数据是否正确收集和返回

## 更新日志

- **v1.0.0**: 初始版本
  - 添加用户资源收集功能
  - 支持进程聚合和TOP进程展示
  - 前端UI集成展示
  - 可配置开关控制