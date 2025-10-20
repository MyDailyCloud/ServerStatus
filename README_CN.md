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

## 📋 开发任务清单

以下任务可供团队成员认领，欢迎在 Issue 中标注认领状态。

### 🔄 v2.2.0 当前版本任务

#### Service 层完善（优先级：高）

- [ ] **WebSocket 服务重构** `data-server/internal/service/websocket.go`
  - 实现 WebSocket 连接管理
  - 实现实时数据推送机制
  - 实现客户端连接池
  - 添加消息队列支持
  - 单元测试覆盖率 ≥ 80%
  - 预计工作量：3-5 天

- [ ] **服务器管理服务** `data-server/internal/service/server_service.go`
  - 实现服务器 CRUD 操作
  - 实现服务器状态管理
  - 实现服务器分组功能
  - 单元测试覆盖率 ≥ 80%
  - 预计工作量：2-3 天

- [ ] **数据导出服务** `data-server/internal/service/export_service.go`
  - 实现 CSV 导出功能
  - 实现 JSON 导出功能
  - 实现批量导出
  - 添加导出任务队列
  - 单元测试覆盖率 ≥ 80%
  - 预计工作量：2-3 天

- [ ] **认证授权服务** `data-server/internal/service/auth_service.go`
  - 实现双密钥认证逻辑
  - 实现访问令牌生成与验证
  - 实现权限检查机制
  - 单元测试覆盖率 ≥ 80%
  - 预计工作量：2-3 天

- [ ] **健康检查服务** `data-server/internal/service/health_service.go`
  - 实现系统健康检查
  - 实现依赖服务检查（Redis、SQLite）
  - 实现性能指标收集
  - 单元测试覆盖率 ≥ 80%
  - 预计工作量：1-2 天

#### Handler 层实现（优先级：高）

- [ ] **HTTP 路由器实现** `data-server/internal/handler/router.go`
  - 设计 RESTful API 路由
  - 实现路由注册机制
  - 添加路由分组功能
  - 预计工作量：1-2 天

- [ ] **中间件系统** `data-server/internal/handler/middleware/`
  - 实现认证中间件
  - 实现日志中间件
  - 实现 CORS 中间件
  - 实现限流中间件
  - 实现错误恢复中间件
  - 预计工作量：2-3 天

- [ ] **API 处理器实现** `data-server/internal/handler/`
  - 实现服务器 API 处理器
  - 实现认证 API 处理器
  - 实现数据导出 API 处理器
  - 实现健康检查 API 处理器
  - 预计工作量：3-4 天

- [ ] **错误处理机制** `data-server/internal/handler/error.go`
  - 统一错误响应格式
  - 实现错误码定义
  - 实现错误日志记录
  - 预计工作量：1 天

- [ ] **请求验证与响应格式化** `data-server/internal/handler/validator.go`
  - 实现请求参数验证
  - 实现响应数据格式化
  - 添加数据脱敏功能
  - 预计工作量：1-2 天

#### 测试与文档（优先级：中）

- [ ] **单元测试编写**
  - Repository 层测试补充
  - Service 层测试（目标覆盖率 80%）
  - Handler 层测试（目标覆盖率 70%）
  - 预计工作量：5-7 天

- [ ] **集成测试实现** `data-server/test/integration/`
  - API 集成测试
  - WebSocket 集成测试
  - 数据库集成测试
  - 缓存集成测试
  - 预计工作量：3-5 天

- [ ] **性能测试** `data-server/test/benchmark/`
  - API 性能基准测试
  - WebSocket 并发测试
  - 缓存性能测试
  - 数据库查询性能测试
  - 预计工作量：2-3 天

- [ ] **API 文档生成**
  - 使用 Swagger/OpenAPI 生成文档
  - 添加 API 使用示例
  - 添加错误码说明
  - 预计工作量：2-3 天

### 🚀 v2.3.0 企业级功能任务

#### 告警系统（优先级：高）

- [ ] **告警规则引擎** `data-server/internal/alert/`
  - 实现阈值告警规则
  - 实现复合条件告警
  - 实现告警级别定义
  - 实现告警静默功能
  - 预计工作量：5-7 天

- [ ] **告警通知渠道**
  - 邮件通知实现
  - Webhook 通知实现
  - 钉钉机器人集成
  - 企业微信机器人集成
  - Slack 集成
  - 预计工作量：5-7 天

- [ ] **告警历史与统计**
  - 告警历史记录
  - 告警统计报表
  - 告警趋势分析
  - 预计工作量：3-4 天

#### 用户权限管理（优先级：高）

- [ ] **用户管理系统** `data-server/internal/user/`
  - 用户 CRUD 操作
  - 用户角色定义
  - 用户组管理
  - 预计工作量：3-5 天

- [ ] **权限控制系统** `data-server/internal/rbac/`
  - RBAC 权限模型实现
  - 资源权限定义
  - 权限检查中间件
  - 预计工作量：4-6 天

- [ ] **审计日志** `data-server/internal/audit/`
  - 操作日志记录
  - 审计日志查询
  - 审计报表生成
  - 预计工作量：2-3 天

#### 数据备份与恢复（优先级：中）

- [ ] **自动备份系统** `data-server/internal/backup/`
  - 定时备份任务
  - 增量备份支持
  - 备份文件管理
  - 预计工作量：3-4 天

- [ ] **数据恢复功能**
  - 备份文件恢复
  - 数据一致性检查
  - 恢复进度显示
  - 预计工作量：2-3 天

#### 多数据库支持（优先级：中）

- [ ] **PostgreSQL 支持** `data-server/internal/repository/postgres/`
  - PostgreSQL Repository 实现
  - 数据迁移脚本
  - 性能优化
  - 预计工作量：5-7 天

- [ ] **MySQL 支持** `data-server/internal/repository/mysql/`
  - MySQL Repository 实现
  - 数据迁移脚本
  - 性能优化
  - 预计工作量：5-7 天

### ☁️ v2.4.0 云原生功能任务

#### Kubernetes 支持（优先级：高）

- [ ] **Kubernetes Operator** `k8s/operator/`
  - CRD 定义
  - Operator 控制器实现
  - 自动扩缩容支持
  - 预计工作量：10-15 天

- [ ] **Helm Chart** `deploy/helm/`
  - Chart 模板编写
  - 配置参数定义
  - 部署文档
  - 预计工作量：3-5 天

#### 可观测性集成（优先级：高）

- [ ] **Prometheus 集成** `data-server/internal/metrics/`
  - Metrics 导出器实现
  - 自定义指标定义
  - Grafana Dashboard
  - 预计工作量：5-7 天

- [ ] **分布式追踪** `data-server/internal/tracing/`
  - OpenTelemetry 集成
  - Jaeger/Zipkin 支持
  - 追踪数据可视化
  - 预计工作量：5-7 天

- [ ] **Grafana 仪表板**
  - 系统监控仪表板
  - 性能分析仪表板
  - 告警仪表板
  - 预计工作量：3-5 天

#### 服务网格支持（优先级：低）

- [ ] **Istio 集成**
  - 服务网格配置
  - 流量管理
  - 安全策略
  - 预计工作量：7-10 天

### 🔧 持续改进任务

#### 性能优化（优先级：中）

- [ ] **API 性能优化**
  - 数据库查询优化
  - 缓存策略优化
  - 并发处理优化
  - 预计工作量：持续进行

- [ ] **WebSocket 性能优化**
  - 连接池优化
  - 消息队列优化
  - 内存使用优化
  - 预计工作量：持续进行

#### 代码质量（优先级：中）

- [ ] **代码重构**
  - 消除代码重复
  - 提高代码可读性
  - 优化代码结构
  - 预计工作量：持续进行

- [ ] **静态代码分析**
  - golangci-lint 规则完善
  - 代码安全扫描
  - 依赖漏洞扫描
  - 预计工作量：持续进行

#### 文档完善（优先级：中）

- [ ] **开发者文档**
  - 架构设计文档
  - API 开发指南
  - 贡献者指南
  - 预计工作量：持续进行

- [ ] **用户文档**
  - 部署指南
  - 配置指南
  - 故障排查指南
  - 预计工作量：持续进行

---

### 📝 任务认领流程

1. 在 [GitHub Issues](https://github.com/MyDailyCloud/ServerStatus/issues) 中找到对应任务
2. 评论表明认领意向，说明预计完成时间
3. Fork 项目并创建功能分支
4. 完成开发并提交 Pull Request
5. 代码审查通过后合并到主分支

### 🎯 任务优先级说明

- **高优先级**：影响核心功能，需要优先完成
- **中优先级**：重要但不紧急，可以按计划推进
- **低优先级**：锦上添花的功能，可以延后处理

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
