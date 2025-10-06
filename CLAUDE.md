# ServerStatus 架构重构规划

## 项目概述
ServerStatus是一个服务器状态监控系统，正在进行从原有架构到Clean Architecture的全面重构。本规划文档详细描述了重构的各个阶段、技术要求、进度跟踪和开发计划。

## 📋 整体进度概览

### ✅ 第一阶段：基础架构 (100% 完成)
- [x] 创建新的目录结构
- [x] 定义核心接口和模型
- [x] 实现配置管理模块
- [x] 基础中间件和工具

### ✅ 第二阶段：Repository层 (100% 完成)
- [x] SQLite Repository实现
- [x] Redis Repository实现
- [x] 数据库连接池管理
- [x] 缓存策略实现
- [x] 数据迁移脚本
- [ ] Repository层单元测试

### 🔄 第三阶段：Service层 (90% 进行中)
- [x] 服务器管理服务 ✅
- [x] 数据导出服务 ✅
- [x] 认证授权服务 ✅
- [x] 健康检查服务 ✅
- [🔄] WebSocket服务 (进行中)
- [ ] Service层单元测试

### ⏳ 第四阶段：Handler层 (待开始)
- [ ] RESTful API Handler
- [ ] WebSocket Handler
- [ ] 中间件集成
- [ ] 请求验证和错误处理

### ⏳ 第五阶段：WebSocket模块重构 (待开始)
- [ ] WebSocket连接管理
- [ ] 实时数据推送
- [ ] 客户端管理
- [ ] 消息路由

### ⏳ 测试和部署 (待开始)
- [ ] 单元测试覆盖率达到60%
- [ ] 集成测试
- [ ] Docker部署文档
- [ ] 安装部署脚本

---

## 🏗️ 详细技术规划

### 第一阶段：基础架构 (100% ✅)

#### 1.1 目录结构重构
```
data-server/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── models/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   └── middleware/
├── pkg/
│   ├── config/
│   ├── logger/
│   └── utils/
└── docs/
```

#### 1.2 核心接口定义
- [x] Repository接口层
- [x] Service接口层
- [x] Handler接口层
- [x] 中间件接口

#### 1.3 配置管理系统
- [x] YAML/JSON配置文件支持
- [x] 环境变量覆盖
- [x] 配置热重载
- [x] 配置验证

### 第二阶段：Repository层 (100% ✅)

#### 2.1 数据库抽象层
- [x] ServerRepository: 服务器数据CRUD
- [x] HistoryRepository: 历史数据管理
- [x] CacheRepository: 缓存抽象接口
- [x] AccessKeyRepository: 访问密钥管理

#### 2.2 数据库实现
- [x] SQLite实现 (主数据库)
- [x] Redis实现 (缓存层)
- [x] 连接池管理
- [x] 事务支持

#### 2.3 数据迁移
- [x] 版本化迁移脚本
- [x] 向前/向后兼容
- [x] 数据备份机制

### 第三阶段：Service层 (90% 🔄)

#### 3.1 核心业务服务
- [x] **服务器管理服务** (`internal/service/server/`)
  - 服务器注册和状态更新
  - 在线状态监控
  - 服务器分组管理
  - 数据聚合和统计

- [x] **数据导出服务** (`internal/service/export/`)
  - CSV格式导出
  - 自定义时间范围
  - 数据过滤和筛选
  - 批量导出支持

- [x] **认证授权服务** (`internal/service/auth/`)
  - 访问密钥验证
  - 权限管理
  - 会话管理
  - 速率限制

- [x] **健康检查服务** (`internal/service/health/`)
  - 系统组件健康监控
  - 自定义健康检查
  - 性能指标收集
  - 告警机制

- [🔄] **WebSocket服务** (`internal/service/websocket/`)
  - 连接管理
  - 实时数据推送
  - 消息路由
  - 客户端认证

#### 3.2 服务特性
- 依赖注入
- 接口抽象
- 错误处理
- 日志集成
- 配置管理

### 第四阶段：Handler层 (⏳ 待开始)

#### 4.1 RESTful API Handler
- [ ] 服务器管理API
- [ ] 数据查询API
- [ ] 配置管理API
- [ ] 健康检查API

#### 4.2 WebSocket Handler
- [ ] 连接升级处理
- [ ] 消息格式验证
- [ ] 客户端认证
- [ ] 错误处理

#### 4.3 中间件
- [ ] 认证中间件
- [ ] 日志中间件
- [ ] 限流中间件
- [ ] CORS中间件
- [ ] 错误处理中间件

### 第五阶段：WebSocket模块重构 (⏳ 待开始)

#### 5.1 连接管理
- [ ] 连接池管理
- [ ] 心跳检测
- [ ] 重连机制
- [ ] 连接状态监控

#### 5.2 实时数据推送
- [ ] 服务器状态推送
- [ ] 告警信息推送
- [ ] 系统事件推送
- [ ] 自定义事件支持

#### 5.3 客户端管理
- [ ] 客户端注册
- [ ] 权限验证
- [ ] 消息过滤
- [ ] 客户端分组

---

## 🧪 测试策略

### 单元测试 (目标覆盖率: 60%)
- [ ] Repository层测试
- [ ] Service层测试
- [ ] Handler层测试
- [ ] 中间件测试
- [ ] 工具函数测试

### 集成测试
- [ ] API集成测试
- [ ] WebSocket集成测试
- [ ] 数据库集成测试
- [ ] 缓存集成测试

### 性能测试
- [ ] 并发连接测试
- [ ] 数据库性能测试
- [ ] 内存使用测试
- [ ] 响应时间测试

---

## 📦 部署规划

### Docker化
- [ ] 多阶段构建
- [ ] 镜像优化
- [ ] 健康检查
- [ ] 配置管理

### 文档
- [ ] API文档
- [ ] 部署文档
- [ ] 配置说明
- [ ] 故障排除指南

### 脚本
- [ ] 安装脚本
- [ ] 部署脚本
- [ ] 备份脚本
- [ ] 监控脚本

---

## 🔧 技术栈

### 核心技术
- **语言**: Go 1.21+
- **数据库**: SQLite (主), Redis (缓存)
- **通信**: HTTP, WebSocket
- **配置**: YAML, 环境变量

### 依赖库
- `github.com/gorilla/mux` - HTTP路由
- `github.com/gorilla/websocket` - WebSocket支持
- `github.com/sirupsen/logrus` - 日志
- `github.com/stretchr/testify` - 测试
- `github.com/spf13/viper` - 配置管理
- `github.com/go-redis/redis` - Redis客户端

### 开发工具
- **测试覆盖率**: `go test -cover`
- **代码检查**: `golangci-lint`
- **依赖管理**: Go Modules
- **构建**: Docker Multi-stage

---

## 📈 性能指标

### 目标性能
- **并发连接**: 1000+ WebSocket连接
- **响应时间**: API < 100ms
- **数据推送**: 实时 < 50ms
- **内存使用**: < 512MB (1000连接)
- **CPU使用**: < 50% (正常负载)

### 监控指标
- 连接数统计
- 请求响应时间
- 错误率统计
- 系统资源使用
- 数据库性能

---

## 🔄 发布计划

### 当前版本: v2.1.0-alpha (第三阶段)
- [x] Service层基础服务
- [x] 健康检查系统
- [x] 认证授权
- [ ] WebSocket服务

### 下一版本: v2.1.0-beta (第四阶段)
- [ ] Handler层实现
- [ ] RESTful API
- [ ] WebSocket集成
- [ ] 中间件系统

### 最终版本: v2.1.0 (正式版)
- [ ] 完整测试覆盖
- [ ] 文档完善
- [ ] 部署脚本
- [ ] 性能优化

---

## 📝 开发规范

### 代码规范
- 遵循Go官方代码规范
- 使用golangci-lint进行代码检查
- 所有公共接口必须有文档注释
- 错误处理必须完整

### Git规范
- 使用语义化版本控制
- 提交信息遵循Conventional Commits
- 功能分支开发
- 代码审查必须

### 测试规范
- 单元测试覆盖率不低于60%
- 所有公共接口必须有测试
- 集成测试覆盖主要流程
- 性能测试验证关键指标

---

## 🚀 下一步行动

### 当前优先级 (2025-01-07)
1. **完成WebSocket服务实现** - 第三阶段最后一个组件
2. **Service层单元测试** - 确保代码质量
3. **开始Handler层开发** - 第四阶段启动

### 本周目标
- [x] 健康检查服务完成
- [🔄] WebSocket服务核心功能
- [ ] WebSocket服务单元测试
- [ ] Service层测试覆盖率达标

### 本月目标
- [ ] 完成第三阶段所有工作
- [ ] 启动第四阶段Handler层
- [ ] 建立CI/CD流程
- [ ] 准备第一个Beta版本

---

*最后更新: 2025-01-07*
*当前状态: 第三阶段Service层 - WebSocket服务开发中*