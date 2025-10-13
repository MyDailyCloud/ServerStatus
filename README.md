# ServerStatus - 轻量级服务器监控系统

[![GitHub Release](https://img.shields.io/github/v/release/MyDailyCloud/ServerStatus)](https://github.com/MyDailyCloud/ServerStatus/releases)
[![License](https://img.shields.io/github/license/MyDailyCloud/ServerStatus)](LICENSE)

一个基于 Go 语言开发的轻量级服务器监控解决方案，支持实时数据采集、多平台部署和灵活的认证机制。

## 📋 目录

- [项目概述](#项目概述)
- [功能特性](#功能特性)
- [系统要求](#系统要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 文档](#api-文档)
- [使用场景](#使用场景)
- [验证测试](#验证测试)
- [故障排查](#故障排查)
- [贡献](#贡献)

## 项目概述

ServerStatus 是一个轻量级的服务器监控系统，提供实时的系统资源监控和数据可视化功能。它采用客户端-服务器架构，通过 WebSocket 实现实时数据推送，支持多种认证方式和项目隔离。

### 主要特点

- 🚀 **轻量高效**: 基于 Go 语言开发，资源占用低，性能优异
- 🔄 **实时监控**: WebSocket 实时数据推送，默认 5 秒更新间隔
- 🔐 **灵活认证**: 支持公开模式、API 密钥和双密钥认证
- 🌍 **跨平台**: 支持 Linux、macOS、Windows 多平台部署
- 📊 **完整监控**: CPU、内存、磁盘、网络、GPU、温度等全面监控
- 💾 **数据持久化**: SQLite 数据库存储，可选 Redis 缓存

## 功能特性

### 系统监控
- ✅ CPU 使用率监控
- ✅ 内存使用率监控
- ✅ 磁盘使用率监控
- ✅ 网络流量统计
- ✅ GPU 信息采集（支持 NVIDIA）
- ✅ 系统温度监控

### 数据传输
- ✅ WebSocket 实时数据推送
- ✅ HTTP RESTful API 接口
- ✅ 自动重连机制
- ✅ 数据压缩传输

### 认证与安全
- ✅ 公开模式（无认证）
- ✅ API 密钥认证
- ✅ 双密钥认证（服务器密钥 + 项目密钥）
- ✅ 访问密钥生成与管理
- ✅ 项目级数据隔离

### 其他功能
- ✅ 多项目支持
- ✅ Session ID 管理
- ✅ 数据保留策略
- ✅ Redis 缓存（可选）
- ✅ 详细的日志记录

## 系统要求

### 服务器端
- 操作系统: Linux、macOS 或 Windows
- 架构: x86_64 (amd64) 或 ARM64
- 磁盘空间: 至少 50MB
- 内存: 建议至少 512MB

### 客户端（监控代理）
- 操作系统: Linux、macOS 或 Windows
- 架构: x86_64 (amd64) 或 ARM64
- 磁盘空间: 至少 20MB
- 内存: 建议至少 128MB

## 安装

### 下载预编译二进制文件

从 [GitHub Releases](https://github.com/MyDailyCloud/ServerStatus/releases) 下载最新版本的预编译二进制文件。

**最新版本**: v1.0.4

#### Linux (x86_64)
```bash
# 下载服务器端
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-linux-amd64
chmod +x data-server-linux-amd64

# 下载监控代理
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-linux-amd64
chmod +x monitor-agent-linux-amd64
```

#### Linux (ARM64)
```bash
# 下载服务器端
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-linux-arm64
chmod +x data-server-linux-arm64

# 下载监控代理
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-linux-arm64
chmod +x monitor-agent-linux-arm64
```

#### macOS (Intel)
```bash
# 下载服务器端
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-darwin-amd64
chmod +x data-server-darwin-amd64

# 下载监控代理
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-darwin-amd64
chmod +x monitor-agent-darwin-amd64
```

#### macOS (Apple Silicon)
```bash
# 下载服务器端
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-darwin-arm64
chmod +x data-server-darwin-arm64

# 下载监控代理
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-darwin-arm64
chmod +x monitor-agent-darwin-arm64
```

#### Windows (x86_64)
下载以下文件并解压到合适的目录：
- [data-server-windows-amd64.exe](https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-windows-amd64.exe)
- [monitor-agent-windows-amd64.exe](https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-windows-amd64.exe)

### 从源码编译

如果需要从源码编译，请参考项目根目录下的构建脚本：
- `scripts/build.sh` - Linux 和 Windows 构建
- `scripts/build-multiplatform.sh` - 多平台构建

## 快速开始

### 基础使用（公开模式）

这是最简单的使用方式，适合测试和个人使用。

#### 1. 启动服务器

```bash
./data-server-linux-amd64 -key public -port 8080
```

服务器将在 `0.0.0.0:8080` 上启动，你将看到类似如下的输出：

```
2025/10/13 16:26:07 启动 ServerStatus Monitor Data Server...
2025/10/13 16:26:07 端口: 8080
2025/10/13 16:26:07 数据限制: 1000 条记录
2025/10/13 16:26:07 推荐数据间隔: 5 秒
2025/10/13 16:26:07 API认证: 禁用
2025/10/13 16:26:07 服务器启动在 0.0.0.0:8080
2025/10/13 16:26:07 访问 http://localhost:8080 查看监控界面
```

#### 2. 启动监控代理（在另一个终端）

```bash
./monitor-agent-linux-amd64 -url http://localhost:8080/api/data -key public
```

代理将开始向服务器报告数据，你将看到类似如下的输出：

```
2025/10/13 16:26:30 启动 ServerStatus Monitor Agent...
2025/10/13 16:26:30 主机名: your-hostname
2025/10/13 16:26:30 上报地址: http://localhost:8080/api/data
2025/10/13 16:26:30 Session注册成功: bf94d98f-b27e-4f9a-8256-9fc06abf9865
2025/10/13 16:26:30 上报间隔: 5s
2025/10/13 16:26:31 成功上报数据 - CPU: 0.4%, 内存: 2.9%, 磁盘: 8.4%, GPU: 无GPU
```

#### 3. 访问监控面板

打开浏览器访问：
```
http://localhost:8080/?key=public
```

或者通过 API 查看数据：
```bash
curl http://localhost:8080/api/servers
```

### 双密钥认证模式

双密钥认证提供更高的安全性和项目隔离能力，适合团队和生产环境使用。

#### 1. 启动服务器（双密钥模式）

```bash
./data-server-linux-amd64 -server-key your-server-secret -port 8080
```

#### 2. 启动监控代理（双密钥模式）

```bash
./monitor-agent-linux-amd64 \
  -url http://localhost:8080/api/data \
  -key project-alpha \
  -server-key your-server-secret
```

代理启动后会自动生成访问链接：

```
=== 🌐 访问链接 ===
📊 项目监控面板: http://localhost:8080/?key=project-alpha
🔐 访问密钥链接: http://localhost:8080/?access=79674874adb7b8bcd35a8e1b3386219fa2edf29e4862a4aacd904bdfe31321ab
```

#### 3. 生成访问密钥（可选）

你也可以手动生成访问密钥：

```bash
curl -X POST http://localhost:8080/api/generate-access-key \
  -H 'Content-Type: application/json' \
  -d '{"server_key": "your-server-secret", "project_key": "project-alpha"}'
```

响应示例：
```json
{
  "access_key": "79674874adb7b8bcd35a8e1b3386219fa2edf29e4862a4aacd904bdfe31321ab",
  "access_url": "http://localhost:8080/?access=79674874adb7b8bcd35a8e1b3386219fa2edf29e4862a4aacd904bdfe31321ab"
}
```

## 配置说明

### 服务器端配置

#### 命令行参数

```bash
./data-server-linux-amd64 [选项]
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-key` | 项目密钥（用于生成访问令牌） | 无 |
| `-server-key` | 服务器密钥（用于双密钥认证） | 无 |
| `-host` | 服务器绑定IP地址 | 0.0.0.0 |
| `-port` | 服务器端口 | 8080 |
| `-config` | 配置文件路径 | server-config.json |
| `-auth` | 启用API密钥认证 | false |
| `-data-limit` | 每台客户端数据保留条数限制 | 1000 |
| `-data-interval` | 推荐的数据上报间隔秒数 | 5 |
| `-help` | 显示帮助信息 | - |

#### 配置文件示例 (server-config.json)

```json
{
  "project_key": "your-project-key",
  "allowed_keys": [
    "key1",
    "key2",
    "key3"
  ],
  "server_key": "your-server-secret",
  "host": "0.0.0.0",
  "port": "8080",
  "require_auth": true,
  "data_limit": 1000,
  "data_interval": 5
}
```

### 客户端配置

#### 命令行参数

```bash
./monitor-agent-linux-amd64 [选项]
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-url` | 服务器上报URL | 无（必需） |
| `-key` | API认证密钥/项目密钥 | 无 |
| `-server-key` | 服务器密钥（双密钥认证） | 无 |
| `-config` | 配置文件路径 | config.json |
| `-silent` | 静默模式（第一次上报成功后不再打印） | false |
| `-help` | 显示帮助信息 | - |

#### 配置文件示例 (config.json)

```json
{
  "server_url": "http://your-server:8080/api/data",
  "project_key": "project-alpha",
  "server_key": "your-server-secret",
  "report_interval": "5s",
  "timeout": "10s"
}
```

#### 使用环境变量

为了安全起见，建议使用环境变量设置敏感信息：

```bash
export SERVER_KEY=your-server-secret
./monitor-agent-linux-amd64 -url http://localhost:8080/api/data -key project-alpha -server-key $SERVER_KEY
```

## API 文档

### 数据上报 API

#### POST /api/data
提交监控数据

**请求头:**
- `Content-Type: application/json`
- `X-API-Key: your-api-key` (可选，启用认证时需要)

**请求体:**
```json
{
  "hostname": "server-01",
  "session_id": "uuid-string",
  "project_key": "project-alpha",
  "cpu_percent": 25.5,
  "memory_percent": 45.2,
  "disk_percent": 60.0,
  "os": "ubuntu",
  "cpu_temp": 55.0,
  "gpu_temp": 45.0,
  "gpus": [...],
  "max_temp": 55.0
}
```

#### POST /api/register-session
注册新的 Session，获取 UUID

**请求体:**
```json
{
  "hostname": "server-01",
  "project_key": "project-alpha"
}
```

**响应:**
```json
{
  "session_id": "bf94d98f-b27e-4f9a-8256-9fc06abf9865",
  "report_interval": 5
}
```

### 数据查询 API

#### GET /api/servers
获取所有服务器列表

**响应示例:**
```json
[
  {
    "hostname": "server-01",
    "session_id": "bf94d98f-b27e-4f9a-8256-9fc06abf9865",
    "last_seen": "2025-10-13T16:27:31Z",
    "status": "online",
    "cpu_percent": 0.4,
    "memory_percent": 2.9,
    "disk_percent": 8.4,
    "os": "ubuntu",
    "cpu_temp": 0,
    "gpu_temp": 0,
    "gpus": null,
    "max_temp": 0
  }
]
```

#### GET /api/server/{hostname}
获取特定服务器详情

#### GET /api/uuid-count
获取统计信息

**响应示例:**
```json
{
  "active_uuids": 1,
  "description": "使用我们服务的设备统计",
  "hostname_only": 0,
  "timestamp": "2025-10-13T16:27:36Z",
  "total_servers": 1
}
```

### 双密钥认证 API

#### POST /api/generate-access-key
生成访问密钥

**请求体:**
```json
{
  "server_key": "your-server-secret",
  "project_key": "project-alpha"
}
```

**响应:**
```json
{
  "access_key": "generated-access-key-hash",
  "access_url": "http://server:8080/?access=generated-access-key-hash"
}
```

#### GET /api/access/{accessKey}/servers
使用访问密钥获取服务器列表

#### GET /api/access/{accessKey}/server/{hostname}
使用访问密钥获取特定服务器详情

#### GET /api/access/{accessKey}/server-by-session/{sessionID}
使用访问密钥和 Session ID 获取服务器详情

### 完整 API 文档

更多 API 详情请参考：[API.md](data-server/API.md)

## 使用场景

### 场景 1: 个人服务器监控

适合个人开发者监控自己的服务器。

```bash
# 服务器端
./data-server-linux-amd64 -key personal -port 8080

# 监控代理（可部署在多台服务器）
./monitor-agent-linux-amd64 -url http://monitor-server:8080/api/data -key personal
```

访问: `http://monitor-server:8080/?key=personal`

### 场景 2: 团队项目监控

不同团队使用不同的项目密钥，实现数据隔离。

```bash
# 服务器端（设置服务器密钥）
./data-server-linux-amd64 -server-key company-secret -port 8080

# 团队 A 的服务器
./monitor-agent-linux-amd64 \
  -url http://monitor-server:8080/api/data \
  -key team-a \
  -server-key company-secret

# 团队 B 的服务器
./monitor-agent-linux-amd64 \
  -url http://monitor-server:8080/api/data \
  -key team-b \
  -server-key company-secret
```

- 团队 A 访问: `http://monitor-server:8080/?key=team-a`
- 团队 B 访问: `http://monitor-server:8080/?key=team-b`

### 场景 3: 生产环境部署

启用认证，使用访问密钥控制访问权限。

```bash
# 服务器端（启用认证）
./data-server-linux-amd64 \
  -auth \
  -server-key production-secret \
  -port 8080

# 监控代理
./monitor-agent-linux-amd64 \
  -url http://monitor-server:8080/api/data \
  -key production \
  -server-key production-secret
```

生成访问密钥供外部访问：
```bash
curl -X POST http://monitor-server:8080/api/generate-access-key \
  -H 'Content-Type: application/json' \
  -d '{"server_key": "production-secret", "project_key": "production"}'
```

### 场景 4: 使用配置文件

对于复杂的部署场景，推荐使用配置文件。

**server-config.json:**
```json
{
  "project_key": "production",
  "server_key": "production-secret",
  "host": "0.0.0.0",
  "port": "8080",
  "require_auth": true,
  "data_limit": 2000,
  "data_interval": 5
}
```

**config.json:**
```json
{
  "server_url": "http://monitor-server:8080/api/data",
  "project_key": "production",
  "server_key": "production-secret",
  "report_interval": "5s",
  "timeout": "10s"
}
```

启动命令：
```bash
# 服务器端
./data-server-linux-amd64 -config server-config.json

# 监控代理
./monitor-agent-linux-amd64 -config config.json
```

## 验证测试

### 验证服务器运行状态

```bash
# 检查服务器是否响应
curl http://localhost:8080/api/uuid-count

# 预期输出
{
  "active_uuids": 1,
  "total_servers": 1,
  ...
}
```

### 验证数据采集

```bash
# 查看所有服务器
curl http://localhost:8080/api/servers

# 预期输出
[
  {
    "hostname": "your-hostname",
    "status": "online",
    "cpu_percent": 0.4,
    "memory_percent": 2.9,
    "disk_percent": 8.4,
    ...
  }
]
```

### 验证数据更新

等待几秒后再次查询，确认数据有更新：

```bash
# 多次执行，观察 last_seen 时间戳变化
curl http://localhost:8080/api/servers | grep last_seen
```

### 验证双密钥认证

```bash
# 生成访问密钥
curl -X POST http://localhost:8080/api/generate-access-key \
  -H 'Content-Type: application/json' \
  -d '{"server_key": "your-server-secret", "project_key": "your-project-key"}'

# 使用访问密钥查询
curl "http://localhost:8080/api/access/{返回的access_key}/servers"
```

## 故障排查

### 服务器无法启动

**问题**: `bind: address already in use`

**解决方案**: 端口被占用，更换端口或关闭占用该端口的进程
```bash
# 检查端口占用
lsof -i :8080

# 使用其他端口
./data-server-linux-amd64 -port 9090
```

### 代理无法连接服务器

**问题**: `connection refused` 或 `timeout`

**解决方案**:
1. 确认服务器地址和端口正确
2. 检查防火墙设置
3. 确认服务器正在运行

```bash
# 测试服务器连通性
curl http://server-ip:8080/api/uuid-count
```

### 数据未更新

**问题**: 数据长时间不更新

**解决方案**:
1. 检查代理日志，确认是否成功上报
2. 检查服务器日志，确认是否收到数据
3. 验证 Session ID 是否正确

### 认证失败

**问题**: `authentication failed` 或 `invalid key`

**解决方案**:
1. 确认项目密钥正确
2. 双密钥模式下，确认服务器密钥一致
3. 检查配置文件中的密钥设置

### 二进制文件无法执行

**问题**: `Exec format error` 或 `cannot execute binary file`

**解决方案**:
1. 确认下载了正确架构的二进制文件（x86_64 vs ARM64）
2. 确认操作系统匹配（Linux vs macOS vs Windows）
3. 确认文件有执行权限 (`chmod +x filename`)

```bash
# 查看系统架构
uname -m

# x86_64 使用 amd64 版本
# aarch64 使用 arm64 版本
```

## 贡献

欢迎贡献代码、报告问题或提出建议！

- 项目地址: [https://github.com/MyDailyCloud/ServerStatus](https://github.com/MyDailyCloud/ServerStatus)
- 问题反馈: [Issues](https://github.com/MyDailyCloud/ServerStatus/issues)
- 功能建议: [Discussions](https://github.com/MyDailyCloud/ServerStatus/discussions)

如果觉得这个项目对你有帮助，欢迎给个 ⭐ Star 支持一下！

## 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

## 相关文档

- [API 完整文档](data-server/API.md)
- [项目路线图](ROADMAP.md)
- [开发指南](CLAUDE.md)

---

**项目地址**: [https://github.com/MyDailyCloud/ServerStatus](https://github.com/MyDailyCloud/ServerStatus)

如有问题或建议，欢迎提 Issue 或 PR！
