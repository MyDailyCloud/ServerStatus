# 架构/开发/使用 反思备忘

目标：保持“清晰、可维护、可观察、易用”。

## 现状概览
- 单一入口：`cmd/server/main.go` → `app.Build` 装配 → `handler.BuildHTTPHandler` 构建中间件/路由。
- 中间件集中：RequestID → Recovery → Logging → BodyLimit → CORS → JSON 404/405。
- Handler 分层：health/server/export/websocket/config/auth，职责单一，依赖接口。
- 文档同步：`docs/flow-handler.md` 直观呈现调用链。

## 设计合理性
- 依赖注入单点，避免隐式耦合，新增功能只改一处（router 构建）。
- Handler 只做输入输出映射，业务落在 Service，数据落在 Repository，调用深度短。
- 中间件顺序明确，可按需开关（CORS/BodyLimit），不阻碍基础路径。
- RequestID + Logging 携带 req_id，便于端到端排障。
- 统一 JSON 404/405，配合 Logging/RequestID，便于排障与可观测。

## 后续开发简化建议
- 统一错误/校验：可新增 validation/error 中间件，在 router 构建处集中挂载。
- 鉴权/限流：如需，可在中间件层按 HandlerConfig 开关，仍保持集中声明。
- WebSocket 升级：若要提供升级/推送接口，按现有模式新建 handler，避免在路由层混入业务。
- 测试：为 config/export/ws/404/405/RequestID 补集成测试，确保响应格式与错误码一致。
- 清理遗留：继续迁移/删除旧 monolith `main.go` 路由逻辑，保持唯一入口。
- 测试样本：`testdata/api-samples.json` 提供可复用请求样例，便于 TDD/集成测试。

## 使用/运维友好性
- 全部为标准 HTTP + gzip；CORS/体积限制可配置；无额外复杂依赖。
- 请求 ID + 结构化日志，便于链路追踪；健康/WS 观测接口方便自检。

