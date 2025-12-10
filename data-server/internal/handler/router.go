package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	authhandler "github.com/kanshan/ServerStatus/data-server/internal/handler/auth"
	confighandler "github.com/kanshan/ServerStatus/data-server/internal/handler/config"
	exporthandler "github.com/kanshan/ServerStatus/data-server/internal/handler/export"
	healthhandler "github.com/kanshan/ServerStatus/data-server/internal/handler/health"
	"github.com/kanshan/ServerStatus/data-server/internal/handler/middleware"
	serverhandler "github.com/kanshan/ServerStatus/data-server/internal/handler/server"
	websockethandler "github.com/kanshan/ServerStatus/data-server/internal/handler/websocket"
)

// BuildHTTPHandler 依据依赖与配置构建 HTTP 路由。
// 目前聚焦健康检查，后续可逐步挂载服务器、认证、导出等 handler。
func BuildHTTPHandler(deps *HandlerDependencies, cfg *HandlerConfig) (http.Handler, error) {
	if deps == nil {
		return nil, fmt.Errorf("handler dependencies is nil")
	}
	if cfg == nil {
		cfg = DefaultHandlerConfig()
	}

	router := mux.NewRouter()
	adapter := &muxAdapter{router: router}

	// 统一的 404/405 JSON 响应，避免返回 HTML
	router.NotFoundHandler = http.HandlerFunc(jsonNotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(jsonMethodNotAllowed)

	// 基础中间件：panic recovery + 请求日志
	if deps.Logger != nil {
		router.Use(middleware.Recovery(deps.Logger))
		router.Use(middleware.Logging(deps.Logger))
	}
	// 可选限流
	if cfg.EnableRateLimit {
		rl := middleware.NewRateLimiter(cfg.RateLimitRequests, time.Duration(cfg.RateLimitWindowSec)*time.Second)
		router.Use(rl.Middleware())
	}
	// 请求 ID，便于追踪
	router.Use(middleware.RequestID())
	// 写操作要求 application/json
	router.Use(middleware.RequireJSON())

	// 请求体大小限制
	if cfg.MaxRequestBodySize > 0 {
		router.Use(middleware.BodyLimit(cfg.MaxRequestBodySize))
	}

	// CORS
	if cfg.EnableCORS {
		router.Use(middleware.CORS(cfg.AllowedOrigins, cfg.AllowedMethods, cfg.AllowedHeaders))
	}

	// 健康检查 handler
	if deps.HealthService == nil {
		return nil, fmt.Errorf("health service is required")
	}
	health := healthhandler.New(deps.HealthService)
	health.SetupRoutes(adapter)

	// 服务器 handler
	if deps.ServerService != nil {
		serverH := serverhandler.New(deps.ServerService)
		serverH.SetupRoutes(adapter)
	}

	// 导出 handler
	if deps.ExportService != nil {
		exportH := exporthandler.New(deps.ExportService)
		exportH.SetupRoutes(adapter)
	}

	// WebSocket handler（只提供状态查询）
	if deps.WebSocketService != nil {
		wsH := websockethandler.New(deps.WebSocketService)
		wsH.SetupRoutes(adapter)
	}

	// 配置 handler
	if deps.ConfigService != nil {
		cfgH := confighandler.New(deps.ConfigService)
		cfgH.SetupRoutes(adapter)
	}

	// 认证 handler
	if deps.AuthService != nil {
		authH := authhandler.New(deps.AuthService)
		authH.SetupRoutes(adapter)
	}

	// TODO: 挂载其他 handler (Server/Auth/Export/WebSocket/Config) 与中间件

	return router, nil
}

// muxAdapter 适配 mux.Router 到 Handler 定义的 Router 接口。
type muxAdapter struct {
	router *mux.Router
}

func (m *muxAdapter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), methods ...string) {
	route := m.router.HandleFunc(pattern, handler)
	if len(methods) > 0 {
		route.Methods(methods...)
	}
}

func (m *muxAdapter) Handle(pattern string, handler http.Handler) {
	m.router.Handle(pattern, handler)
}

func (m *muxAdapter) Use(middleware ...func(http.Handler) http.Handler) {
	for _, mw := range middleware {
		m.router.Use(mux.MiddlewareFunc(mw))
	}
}

func jsonNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "not_found",
		"message": "resource not found",
		"path":    r.URL.Path,
	})
}

func jsonMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "method_not_allowed",
		"message": "method not allowed",
		"path":    r.URL.Path,
	})
}
