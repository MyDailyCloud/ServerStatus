# ServerStatus Monitor API 文档

## 概述

ServerStatus Monitor 是一个服务器监控系统，提供RESTful API用于数据收集和查询。

**Base URL**: `http://your-server:8080`

**API Version**: v1

## 认证方式

系统支持以下认证方式：

1. **公开访问** - 不需要认证，显示ProjectKey为"public"的服务器
2. **访问密钥认证** - 使用服务器密钥+项目密钥生成的访问密钥

## 数据结构

### SystemInfo - 系统信息上报数据
```json
{
  "hostname": "server-01",
  "session_id": "uuid-string",
  "timestamp": "2024-01-01T00:00:00Z",
  "cpu": {
    "usage_percent": 45.2,
    "core_count": 8,
    "model_name": "Intel Core i7"
  },
  "memory": {
    "total": 17179869184,
    "used": 8589934592,
    "free": 8589934592,
    "usage_percent": 50.0
  },
  "disk": {
    "total": 1000000000000,
    "used": 500000000000,
    "free": 500000000000,
    "usage_percent": 50.0
  },
  "network": {
    "bytes_sent": 1048576,
    "bytes_recv": 2097152,
    "packets_sent": 1000,
    "packets_recv": 1500,
    "speed_sent": 100.5,
    "speed_recv": 200.8,
    "interfaces": [
      {
        "name": "eth0",
        "bytes_sent": 1048576,
        "bytes_recv": 2097152,
        "packets_sent": 1000,
        "packets_recv": 1500,
        "speed_sent": 100.5,
        "speed_recv": 200.8,
        "is_up": true,
        "mtu": 1500,
        "addrs": ["192.168.1.100", "fe80::1"]
      }
    ]
  },
  "gpu": {
    "name": "NVIDIA RTX 4090",
    "memory_total": 25769803776,
    "memory_used": 8589934592,
    "usage_percent": 75.5,
    "temperature": 65.0
  },
  "gpus": [
    {
      "name": "NVIDIA RTX 4090",
      "memory_total": 25769803776,
      "memory_used": 8589934592,
      "usage_percent": 75.5,
      "temperature": 65.0
    }
  ],
  "os": {
    "platform": "linux",
    "version": "Ubuntu 22.04",
    "arch": "amd64",
    "uptime": 86400
  },
  "temperature": {
    "cpu_temp": 45.0,
    "gpu_temp": 65.0,
    "sensors": {
      "Core 0": 42.0,
      "Core 1": 45.0
    },
    "max_temp": 65.0,
    "avg_temp": 52.5
  },
  "project_key": "project-alpha"
}
```

### ServerStatus - 服务器状态列表数据
```json
{
  "hostname": "server-01",
  "session_id": "uuid-string",
  "last_seen": "2024-01-01T00:00:00Z",
  "status": "online",
  "cpu_percent": 45.2,
  "memory_percent": 50.0,
  "disk_percent": 50.0,
  "os": "linux",
  "cpu_temp": 45.0,
  "gpu_temp": 65.0,
  "gpus": [...],
  "max_temp": 65.0,
  "network_speed_sent": 100.5,
  "network_speed_recv": 200.8,
  "network_bytes_sent": 1048576,
  "network_bytes_recv": 2097152
}
```

## API 端点

### 1. 数据上报

#### POST /api/data
接收监控代理上报的系统数据。

**Headers:**
- `Content-Type: application/json`
- `X-Server-Key: string` (可选，启用认证时需要)
- `X-Project-Key: string` (可选，用于数据分组，默认为"default")

**Request Body:** SystemInfo 对象

**Response:**
- `200 OK` - 数据接收成功
- `400 Bad Request` - 请求数据格式错误
- `401 Unauthorized` - 认证失败

**Example:**
```bash
curl -X POST http://localhost:8080/api/data \
  -H "Content-Type: application/json" \
  -H "X-Project-Key: public" \
  -d @system_data.json
```

### 2. Session 注册

#### POST /api/register-session
为监控代理注册新的UUID session。

**Request Body:**
```json
{
  "hostname": "server-01",
  "project_key": "public"
}
```

**Response:**
```json
{
  "session_id": "generated-uuid",
  "hostname": "server-01"
}
```

### 3. 服务器列表查询

#### GET /api/servers
获取所有公开服务器列表（ProjectKey为"public"）。

**Response:** ServerStatus 数组

**Example:**
```bash
curl http://localhost:8080/api/servers
```

#### GET /api/server/{hostname}
获取特定服务器的详细信息和历史数据。

**Parameters:**
- `hostname` - 服务器主机名

**Response:** ServerInfo 对象（包含最新数据和历史数据）

### 4. 访问密钥认证

#### POST /api/generate-access-key
使用服务器密钥和项目密钥生成访问密钥。

**Request Body:**
```json
{
  "server_key": "server-secret-key",
  "project_key": "project-alpha"
}
```

**Response:**
```json
{
  "access_key": "generated-access-key-hash",
  "project_key": "project-alpha",
  "message": "访问密钥生成成功"
}
```

#### GET /api/access/{accessKey}/servers
使用访问密钥获取特定项目的服务器列表。

**Parameters:**
- `accessKey` - 生成的访问密钥

**Response:** ServerStatus 数组

#### GET /api/access/{accessKey}/server/{hostname}
使用访问密钥获取特定服务器详情。

**Parameters:**
- `accessKey` - 访问密钥
- `hostname` - 服务器主机名

**Response:** ServerInfo 对象

#### GET /api/access/{accessKey}/server-by-session/{sessionID}
使用访问密钥和sessionID获取特定服务器详情。

**Parameters:**
- `accessKey` - 访问密钥
- `sessionID` - 服务器Session ID

**Response:** ServerInfo 对象

### 5. 统计信息

#### GET /api/uuid-count
获取UUID设备统计信息。

**Response:**
```json
{
  "total_servers": 100,
  "active_uuids": 85,
  "hostname_only": 15,
  "timestamp": "2024-01-01T00:00:00Z",
  "description": "使用我们服务的设备统计"
}
```

### 6. 文件下载

#### GET /download/{filename}
下载监控代理程序。

**Parameters:**
- `filename` - 文件名（如 "monitor-agent-linux"）

**Response:** 二进制文件

#### GET /install
获取一键安装脚本。

**Response:** Shell脚本文件

## 错误响应格式

所有错误响应都使用标准HTTP状态码，响应体为纯文本错误信息：

```
400 Bad Request - 请求格式错误
401 Unauthorized - 认证失败
403 Forbidden - 访问被拒绝
404 Not Found - 资源不存在
500 Internal Server Error - 服务器内部错误
```

## 使用示例

### 1. 基础监控设置
```bash
# 1. 启动服务器（公开模式）
./data-server -key public

# 2. 客户端上报数据
./monitor-agent -key public
```

### 2. 项目隔离监控
```bash
# 1. 启动服务器（启用认证）
./data-server -server-key "my-server-secret" -auth

# 2. 生成项目访问密钥
curl -X POST http://localhost:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "my-server-secret", "project_key": "project-alpha"}'

# 3. 使用访问密钥查看项目服务器
curl http://localhost:8080/api/access/{access_key}/servers
```

### 3. 自定义前端集成
```javascript
// 获取服务器列表
fetch('/api/servers')
  .then(response => response.json())
  .then(servers => {
    servers.forEach(server => {
      console.log(`${server.hostname}: CPU ${server.cpu_percent}%`);
    });
  });

// 获取特定服务器详情
fetch('/api/server/server-01')
  .then(response => response.json())
  .then(serverInfo => {
    console.log('最新数据:', serverInfo.latest);
    console.log('历史数据:', serverInfo.history);
  });
```

## 配置选项

### 服务器配置
```json
{
  "project_key": "default-project-key",
  "server_key": "server-secret-key", 
  "host": "0.0.0.0",
  "port": "8080",
  "require_auth": false,
  "data_limit": 1000,
  "data_interval": 5
}
```

### 环境变量
- `DATA_LIMIT` - 数据保留条数限制
- `DATA_INTERVAL` - 推荐数据上报间隔（秒）
- `REQUIRE_AUTH` - 是否启用认证

## 数据保留策略

- 每台服务器最多保留 `data_limit` 条历史记录（默认1000条）
- 离线超过10分钟的服务器会被自动清理
- 在线状态判断：最后数据上报时间超过30秒视为离线

## 网络相关字段说明

- `network_speed_sent/recv` - 网络发送/接收速率（KB/s）
- `network_bytes_sent/recv` - 总发送/接收字节数
- `interfaces` - 各网卡详细信息
- 速率计算基于两次上报间的差值，首次上报速率为0

## 扩展开发

本API完全支持前后端分离架构，你可以：

1. 使用任何前端框架（React、Vue、Angular等）开发自定义界面
2. 集成到现有监控系统中
3. 开发移动端应用
4. 创建第三方监控工具

所有数据通过标准REST API获取，支持CORS跨域访问。