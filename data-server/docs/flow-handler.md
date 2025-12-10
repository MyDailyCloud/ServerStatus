# HTTP Handler 流程概览

目标：保持调用链简单、可追踪。

```
cmd/server/main.go
  └─ app.Build()                // 装配配置、日志、仓库、服务
      └─ components.*Service    // Server/Auth/Health/WebSocket
  └─ handler.BuildHTTPHandler   // 构建 mux 路由与中间件
      ├─ middleware.RequestID   // 统一请求ID
      ├─ middleware.Recovery    // panic 保护
      ├─ middleware.Logging     // 请求日志 (含 req_id, ip)
      ├─ middleware.RateLimit?  // 可选，简单按 IP 限流
      ├─ middleware.BodyLimit   // 请求体大小限制
      ├─ middleware.CORS        // CORS 允许列表
      ├─ middleware.RequireJSON // 写操作 Content-Type 校验
      ├─ JSON 404/405 handler   // 统一未找到/方法禁止响应
      ├─ health handler         // /api/health, /api/health/info
      ├─ server handler         // /api/servers, /api/projects/{project_key}/servers
      ├─ websocket handler      // /api/ws/stats, /api/ws/connections
      ├─ export handler         // /api/export/*
      ├─ config handler         // /api/config, /api/config/reload
      └─ auth handler           // /api/auth/validate, /api/auth/generate
  └─ http.Server + gzip         // 监听/超时配置
```

分层职责
- Handler：解析请求、调用对应 Service、格式化 HTTP 响应。
- Service：业务用例实现；不关心 HTTP。
- Repository：数据访问。

扩展建议
- 在 `handler/router.go` 挂载更多 handler（export/websocket/config）与横切中间件（日志、认证、限流、CORS、错误格式化）。
- 保持每条路径的调用深度可见：Handler -> Service -> Repository，不在 Handler 内嵌套复杂分支。***

