# ServerStatus v2.x - 新一代服务器监控系统

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)
[![Release](https://img.shields.io/badge/release-v2.2.0-brightgreen.svg)](https://github.com/MyDailyCloud/ServerStatus/releases)
[![Architecture](https://img.shields.io/badge/architecture-Clean%20Architecture-green.svg)](CLAUDE.md)

企业级分布式服务器监控解决方案——采用清洁架构设计，支持实时监控、多租户隔离、高性能缓存。

中文 | [English](README.md)

---

## 📖 项目简介

ServerStatus v2.x 是一个企业级的轻量化服务器监控系统，采用 Go 语言开发，基于清洁架构（Clean Architecture）设计。相比 v1.x 版本，v2.x 在架构、性能和功能上都有显著提升。

### ✨ v2.x 核心特性

**架构升级**
- 清洁架构分层设计（Handler → Service → Repository → Models）
- 接口抽象与依赖注入
- 模块化组件，易于扩展和维护

**性能提升**
- Redis + 内存双层缓存系统
- API 响应速度提升 50-70%
- WebSocket 支持 1000+ 并发连接
- 数据库索引优化与连接池管理

**功能增强**
- 全面监控：CPU、内存、磁盘、网络、GPU、温度、用户资源
- 多 GPU 支持（NVIDIA 多卡监控）
- WebSocket 实时数据推送
- 多租户数据隔离（基于 ProjectKey）
- 双密钥认证系统（ServerKey + ProjectKey）
- SQLite 历史数据存储
- CSV/JSON 数据导出

**跨平台与部署**
- 支持 Linux、Windows、macOS（x86_64/ARM64）
- Docker/Docker Compose 一键部署
- Kubernetes 部署支持（规划中）

---

## 📊 版本对比（v1.x vs v2.x）

| 特性 | v1.x | v2.x |
|---|---|---|
| **架构设计** | 单体架构 | 清洁架构（分层） |
| **缓存系统** | 无 | Redis + 内存双层缓存 |
| **实时推送** | HTTP 轮询 | WebSocket 实时推送 |
| **多 GPU 支持** | 基础单卡 | 完整多卡监控 |
| **数据持久化** | 无 | SQLite 历史数据存储 |
| **认证系统** | 单密钥 | 双密钥 + 访问令牌 |
| **测试覆盖** | 无 | 60%+（目标） |
| **API 响应** | 基准 | 提升 50-70% |
| **并发连接** | 100+ | 1000+ |
| **配置管理** | 命令行参数 | 配置文件 + 环境变量 + 热重载 |

---

## 🚀 快速开始

### 方式一：自动安装脚本（推荐）

```bash
# 下载安装脚本
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/install.sh -o install.sh
chmod +x install.sh

# 安装服务端（交互式）
./install.sh server

# 安装监控客户端（交互式）
./install.sh client
```

脚本功能：
- 自动检测平台和架构
- 下载对应的二进制文件
- 生成配置文件
- 可选注册为系统服务

### 方式二：手动安装

**服务端部署（示例：Linux x86_64）**

```bash
# 下载服务端
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-linux-amd64
chmod +x data-server-linux-amd64

# 启动服务端
./data-server-linux-amd64 -key public -port 8080
```

**监控客户端部署（在被监控主机上）**

```bash
# 下载客户端
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-linux-amd64
chmod +x monitor-agent-linux-amd64

# 启动客户端
./monitor-agent-linux-amd64 -url http://<服务器IP>:8080/api/data -key public
```

**访问监控面板**

```
http://<服务器IP>:8080/?key=public
```

### 方式三：Docker/Docker Compose

```bash
cd deploy
docker-compose up -d

# 访问面板
http://localhost:8080/?key=public
```

---

## 🏗️ 架构设计与进度

### 清洁架构分层

```
┌─────────────────────────────────────────┐
│              Handler Layer               │  HTTP请求处理、路由、中间件
├─────────────────────────────────────────┤
│               Service Layer              │  业务逻辑、用例实现
├─────────────────────────────────────────┤
│             Repository Layer             │  数据访问、存储抽象
├─────────────────────────────────────────┤
│                Models Layer              │  领域模型、数据结构
└─────────────────────────────────────────┘
```

### 重构进度

- ✅ **Repository 层**：100% 完成
- 🔄 **Service 层**：90% 完成（WebSocket 服务开发中）
- ⏳ **Handler 层**：0%（规划中）

**详细架构文档**：[CLAUDE.md](CLAUDE.md)  
**发展路线图**：[ROADMAP.md](ROADMAP.md)

---

## 🗺️ 发展路线图

### v2.2.0（当前版本）

- ✅ 清洁架构重构
- ✅ Repository 层完成
- 🔄 Service 层开发中
- ⏳ Handler 层规划中

### v2.3.0（企业级功能）

- 告警系统（邮件、Webhook、钉钉、企业微信）
- 用户权限管理
- 审计日志
- 数据备份与恢复
- 多数据库支持（PostgreSQL、MySQL）

### v2.4.0（云原生）

- Kubernetes 原生支持
- Prometheus 集成
- Grafana 仪表板
- 分布式追踪
- 服务网格支持

**完整路线图**：[ROADMAP.md](ROADMAP.md)

---

## 👥 贡献指南

### 开发环境

- Go 1.21+
- Redis 7+
- SQLite 3+

### 提交规范

遵循 Conventional Commits 规范：

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建/工具链相关
```

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

如果该项目对你有帮助，欢迎 ⭐ Star 支持！
