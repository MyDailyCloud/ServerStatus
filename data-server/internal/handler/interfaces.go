package handler

import (
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/internal/service"
)

// Handler HTTP处理器接口
type Handler interface {
	http.Handler
	SetupRoutes(mux Router)
}

// Router 路由器接口
type Router interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
	Use(middleware ...func(http.Handler) http.Handler)
}

// ServerHandler 服务器API处理器接口
type ServerHandler interface {
	Handler
	// 服务器相关API
	GetServers(w http.ResponseWriter, r *http.Request)
	GetServer(w http.ResponseWriter, r *http.Request)
	GetServersByProject(w http.ResponseWriter, r *http.Request)
	DeleteServer(w http.ResponseWriter, r *http.Request)

	// 统计API
	GetStats(w http.ResponseWriter, r *http.Request)
	GetProjectStats(w http.ResponseWriter, r *http.Request)
}

// ExportHandler 导出API处理器接口
type ExportHandler interface {
	Handler
	// 导出API
	ExportServers(w http.ResponseWriter, r *http.Request)
	ExportHistory(w http.ResponseWriter, r *http.Request)
	ExportUserResources(w http.ResponseWriter, r *http.Request)
	GetExportFormats(w http.ResponseWriter, r *http.Request)
	GetExportPreview(w http.ResponseWriter, r *http.Request)
}

// AuthHandler 认证API处理器接口
type AuthHandler interface {
	Handler
	// 认证API
	GenerateAccessKey(w http.ResponseWriter, r *http.Request)
	ValidateAccessKey(w http.ResponseWriter, r *http.Request)
	InvalidateAccessKey(w http.ResponseWriter, r *http.Request)
}

// HealthHandler 健康检查API处理器接口
type HealthHandler interface {
	Handler
	// 健康检查API
	CheckHealth(w http.ResponseWriter, r *http.Request)
	CheckComponentHealth(w http.ResponseWriter, r *http.Request)
	GetVersion(w http.ResponseWriter, r *http.Request)
	GetStats(w http.ResponseWriter, r *http.Request)
	GetMetrics(w http.ResponseWriter, r *http.Request)
}

// ConfigHandler 配置API处理器接口
type ConfigHandler interface {
	Handler
	// 配置API
	GetConfig(w http.ResponseWriter, r *http.Request)
	ReloadConfig(w http.ResponseWriter, r *http.Request)
	UpdateConfig(w http.ResponseWriter, r *http.Request)
	GetConfigHistory(w http.ResponseWriter, r *http.Request)
}

// WebSocketHandler WebSocket处理器接口
type WebSocketHandler interface {
	Handler
	// WebSocket连接
	HandleWebSocket(w http.ResponseWriter, r *http.Request)
	GetWebSocketStats(w http.ResponseWriter, r *http.Request)
}

// Middleware 中间件接口
type Middleware interface {
	Handler() func(http.Handler) http.Handler
}

// CORSMiddleware CORS中间件接口
type CORSMiddleware interface {
	Middleware
	WithOrigins(origins []string) CORSMiddleware
	WithMethods(methods []string) CORSMiddleware
	WithHeaders(headers []string) CORSMiddleware
}

// AuthMiddleware 认证中间件接口
type AuthMiddleware interface {
	Middleware
	WithRequired(required bool) AuthMiddleware
	WithSkipPaths(paths []string) AuthMiddleware
}

// LoggingMiddleware 日志中间件接口
type LoggingMiddleware interface {
	Middleware
	WithLevel(level string) LoggingMiddleware
	WithRequestID(enabled bool) LoggingMiddleware
}

// RateLimitMiddleware 限流中间件接口
type RateLimitMiddleware interface {
	Middleware
	WithRate(requests int, window string) RateLimitMiddleware
	WithBurst(burst int) RateLimitMiddleware
}

// CompressionMiddleware 压缩中间件接口
type CompressionMiddleware interface {
	Middleware
	WithLevel(level int) CompressionMiddleware
	WithTypes(types []string) CompressionMiddleware
}

// MetricsMiddleware 指标收集中间件接口
type MetricsMiddleware interface {
	Middleware
	WithPath(enabled bool) MetricsMiddleware
	WithLatency(enabled bool) MetricsMiddleware
}

// RequestValidator 请求验证器接口
type RequestValidator interface {
	Validate(r *http.Request) error
	GetSchema() interface{}
}

// ResponseFormatter 响应格式化器接口
type ResponseFormatter interface {
	Format(w http.ResponseWriter, r *http.Request, data interface{}) error
	GetContentType() string
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(w http.ResponseWriter, r *http.Request, err error)
	HandlePanic(w http.ResponseWriter, r *http.Request, recovered interface{})
}

// HandlerDependencies 处理器依赖
type HandlerDependencies struct {
	ServerService    service.ServerService
	ExportService    service.ExportService
	AuthService      service.AuthService
	HealthService    service.HealthService
	ConfigService    service.ConfigService
	WebSocketService service.WebSocketService
}

// HandlerConfig 处理器配置
type HandlerConfig struct {
	EnableCORS         bool     `json:"enable_cors"`
	AllowedOrigins     []string `json:"allowed_origins"`
	AllowedMethods     []string `json:"allowed_methods"`
	AllowedHeaders     []string `json:"allowed_headers"`
	EnableAuth         bool     `json:"enable_auth"`
	EnableLogging      bool     `json:"enable_logging"`
	EnableMetrics      bool     `json:"enable_metrics"`
	EnableRateLimit    bool     `json:"enable_rate_limit"`
	EnableCompression  bool     `json:"enable_compression"`
	RequestTimeout     string   `json:"request_timeout"`
	MaxRequestBodySize int64    `json:"max_request_body_size"`
}

// DefaultHandlerConfig 返回默认处理器配置
func DefaultHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		EnableCORS:         true,
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Content-Type", "Authorization", "X-Server-Key", "X-Project-Key"},
		EnableAuth:         true,
		EnableLogging:      true,
		EnableMetrics:      true,
		EnableRateLimit:    false,
		EnableCompression:  true,
		RequestTimeout:     "30s",
		MaxRequestBodySize: 10 << 20, // 10MB
	}
}

// HandlerFactory 处理器工厂接口
type HandlerFactory interface {
	CreateServerHandler(deps *HandlerDependencies, config *HandlerConfig) ServerHandler
	CreateExportHandler(deps *HandlerDependencies, config *HandlerConfig) ExportHandler
	CreateAuthHandler(deps *HandlerDependencies, config *HandlerConfig) AuthHandler
	CreateHealthHandler(deps *HandlerDependencies, config *HandlerConfig) HealthHandler
	CreateConfigHandler(deps *HandlerDependencies, config *HandlerConfig) ConfigHandler
	CreateWebSocketHandler(deps *HandlerDependencies, config *HandlerConfig) WebSocketHandler

	// 中间件工厂
	CreateCORSMiddleware(config *HandlerConfig) CORSMiddleware
	CreateAuthMiddleware(deps *HandlerDependencies, config *HandlerConfig) AuthMiddleware
	CreateLoggingMiddleware(config *HandlerConfig) LoggingMiddleware
	CreateRateLimitMiddleware(config *HandlerConfig) RateLimitMiddleware
	CreateCompressionMiddleware(config *HandlerConfig) CompressionMiddleware
	CreateMetricsMiddleware(config *HandlerConfig) MetricsMiddleware
}

// RouteConfig 路由配置
type RouteConfig struct {
	Pattern     string
	Method      string
	Handler     http.Handler
	Middleware  []func(http.Handler) http.Handler
	Auth        bool
	RateLimit   bool
	Compression bool
}

// RouteGroup 路由组
type RouteGroup struct {
	Prefix     string
	Middleware []func(http.Handler) http.Handler
	Routes     []RouteConfig
}

// RouterBuilder 路由构建器接口
type RouterBuilder interface {
	Group(prefix string) *RouteGroup
	AddRoute(config RouteConfig)
	AddGroup(group *RouteGroup)
	Build() Router
}

// APIVersion API版本
type APIVersion struct {
	Version    string  `json:"version"`
	Path       string  `json:"path"`
	Deprecated bool    `json:"deprecated"`
	SunsetAt   *string `json:"sunset_at"`
}
