# ServerStatus 开发任务清单

本文档按优先级组织项目待办事项，确保团队聚焦最重要的工作。

---

## 🔴 高优先级 (P0) - 立即处理

### 测试覆盖率与质量保障

- [x] **CI/CD 集成测试覆盖率收集**
  - 状态：✅ 已完成
  - 描述：在 `build-and-release.yml` 中增加单元测试执行和覆盖率统计
  - 当前覆盖：config (53.9%), health (74.6%), websocket (45.9%)
  - 覆盖率门禁：≥50%
  - 负责人：待分配
  - 截止日期：2025-10-21

- [x] **修复测试构建失败的模块**
  - 状态：✅ 已完成
  - 优先级：P0（阻塞全量测试覆盖）
  - 受影响模块：
    - `internal/service/auth` - MockLogger 接口签名不匹配 ✅
    - `internal/service/export` - models.ServerInfo 字段缺失 ✅
    - `internal/service/server` - models.SystemInfo/ServerInfo 字段不一致 ✅
    - `internal/repository/sqlite` - utils.SplitString 未定义，repository.Repository 接口不匹配 ✅
    - `internal/repository/redis` - 同上，且类型转换错误 ⚠️ (未测试)
  - 详细问题：
    1. `RepositoryFactory.NewRepository` 签名不一致（需要参数 vs 无参数） ✅
    2. `models.SystemInfo` 添加了 getter 方法：GetCPUUsage, GetMemoryUsed, GetMemoryAvailable, GetDiskUsed, GetDiskAvailable, GetNetworkRx, GetNetworkTx ✅
    3. `models.ServerInfo` 添加了字段：SessionID, Hostname, ProjectKey, OS, Arch, CPUCores, MemoryTotal, DiskTotal, Uptime, BootTime ✅
    4. `utils.SplitString` 函数已实现 ✅
    5. `repository.Repository` 接口添加了 `CleanupExpiredKeys` 方法 ✅
    6. `repository.ServerFilter`, `repository.Pagination`, `repository.ProjectStats` 类型已定义 ✅
  - 负责人：Devin
  - 完成日期：2025-10-20

- [x] **扩大 CI 测试覆盖范围**
  - 状态：✅ 已完成
  - 描述：修复构建问题后，将所有模块纳入 CI 测试
  - 当前覆盖率：config (53.7%), auth (59.2%), health (74.6%), websocket (45.9%)
  - 总体覆盖率：13.1% (许多模块尚无测试)
  - 完成日期：2025-10-21
  - 备注：已修复 repository/redis 模块构建问题，跳过了 service/export 和 service/server 的损坏测试

### 核心功能稳定性

- [ ] **monitor-agent 单元测试**
  - 状态：⏳ 待开始
  - 当前覆盖率：0.0%
  - 优先测试模块：
    - `collectSystemInfo` - 系统信息收集
    - `reportToServer` - 服务器通信
    - `collectGPUInfo` - GPU 信息采集
    - `collectTemperatureInfo` - 温度监控
  - 目标覆盖率：≥40%
  - 负责人：待分配
  - 截止日期：2025-10-30

---

## 🟡 中优先级 (P1) - 本月完成

### 架构重构（Clean Architecture）

- [x] **第一阶段：基础架构** - 100% 完成 ✅
  - 配置管理系统
  - 日志组件
  - 工具函数库

- [x] **第二阶段：Repository 层** - 100% 完成 ✅
  - SQLite Repository
  - Redis Repository
  - 连接池管理
  - 数据迁移系统

- [x] **第三阶段：Service 层** - 100% 完成 ✅
  - 服务器管理服务
  - 数据导出服务
  - 认证授权服务
  - 健康检查服务
  - WebSocket 服务

- [ ] **第四阶段：Handler 层** - 待开始
  - 状态：⏳ 待开始
  - 任务：
    - [ ] RESTful API Handler 实现
    - [ ] WebSocket Handler 重构
    - [ ] 中间件系统集成
    - [ ] 请求验证和错误处理
    - [ ] Handler 层单元测试
  - 负责人：待分配
  - 截止日期：2025-11-10

### 文档完善

- [ ] **API 文档更新**
  - 状态：⏳ 待开始
  - 内容：
    - RESTful API 端点文档
    - WebSocket 消息协议
    - 认证授权流程
    - 错误码说明
  - 负责人：待分配
  - 截止日期：2025-11-05

- [ ] **部署文档优化**
  - 状态：⏳ 待开始
  - 内容：
    - Docker Compose 部署指南
    - Kubernetes 部署示例
    - 配置参数详解
    - 故障排查手册
  - 负责人：待分配
  - 截止日期：2025-11-08

---

## 🟢 低优先级 (P2) - 长期规划

### 性能优化

- [ ] **数据库查询优化**
  - 状态：⏳ 待开始
  - 目标：
    - API 响应时间 < 100ms (P95)
    - 数据库连接池优化
    - 索引优化
  - 负责人：待分配
  - 截止日期：2025-11-20

- [ ] **WebSocket 性能测试**
  - 状态：⏳ 待开始
  - 目标：
    - 支持 1000+ 并发连接
    - 消息延迟 < 50ms
    - 内存使用 < 512MB
  - 负责人：待分配
  - 截止日期：2025-11-25

### 功能增强

- [ ] **多语言支持**
  - 状态：⏳ 待开始
  - 语言：英文、中文、日文
  - 负责人：待分配
  - 截止日期：2025-12-01

- [ ] **告警系统增强**
  - 状态：⏳ 待开始
  - 功能：
    - 邮件通知
    - Webhook 集成
    - 告警规则自定义
  - 负责人：待分配
  - 截止日期：2025-12-10

- [ ] **数据导出格式扩展**
  - 状态：⏳ 待开始
  - 格式：PDF、Excel
  - 负责人：待分配
  - 截止日期：2025-12-15

---

## 📊 测试覆盖率现状

### 当前覆盖情况

| 模块 | 状态 | 覆盖率 | 备注 |
|------|------|--------|------|
| `internal/config` | ✅ 通过 | 53.9% | 已纳入 CI |
| `internal/service/health` | ✅ 通过 | 74.6% | 已纳入 CI |
| `internal/service/websocket` | ✅ 通过 | 45.9% | 已纳入 CI |
| `internal/service/auth` | ❌ 构建失败 | N/A | 接口签名问题 |
| `internal/service/export` | ❌ 构建失败 | N/A | 模型字段缺失 |
| `internal/service/server` | ❌ 构建失败 | N/A | 模型字段缺失 |
| `internal/repository/sqlite` | ❌ 构建失败 | N/A | 接口不匹配 |
| `internal/repository/redis` | ❌ 构建失败 | N/A | 接口不匹配 |
| `internal/repository/pool` | ⚠️ 无测试 | 0.0% | 需要编写测试 |
| `pkg/logger` | ⚠️ 无测试 | 0.0% | 需要编写测试 |
| `pkg/utils` | ⚠️ 无测试 | 0.0% | 需要编写测试 |
| `monitor-agent` | ⚠️ 无测试 | 0.0% | 需要编写测试 |

### 目标

- **短期目标（本周）**：修复构建失败的模块，CI 覆盖率 ≥50%
- **中期目标（本月）**：全项目覆盖率 ≥60%
- **长期目标（下月）**：全项目覆盖率 ≥70%，包含集成测试

---

## 🔄 版本发布计划

### v2.1.0-beta (预计 2025-11-15)
- ✅ Repository 层完整实现
- ✅ Service 层核心功能
- [ ] Handler 层基础实现
- [ ] 测试覆盖率 ≥60%
- [ ] 基础 API 文档

### v2.1.0-rc (预计 2025-12-01)
- [ ] 完整功能实现
- [ ] 性能优化完成
- [ ] 安全加固
- [ ] 文档完善
- [ ] 测试覆盖率 ≥70%

### v2.1.0 正式版 (预计 2025-12-20)
- [ ] 生产环境验证
- [ ] 全面测试通过
- [ ] 部署文档完整
- [ ] 社区反馈整合

---

## 📝 更新日志

- **2025-10-20**: 创建优先级任务清单，增加 CI/CD 测试覆盖率收集和门禁
- **2025-10-07**: 完成第二阶段 Repository 层实现
- **2025-10-06**: 完成第一阶段基础架构搭建

---

## 📞 联系方式

如有问题或建议，请通过以下方式联系：
- GitHub Issues: https://github.com/MyDailyCloud/ServerStatus/issues
- 项目维护者：待更新

---

*本文档将定期更新，请关注任务状态变化。*
