package healthhandler

import (
	"context"
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/internal/handler/respond"
	servicehealth "github.com/kanshan/ServerStatus/data-server/internal/service/health"
)

// Service 定义健康处理器所需的最小服务接口，避免直接依赖具体实现
type Service interface {
	CheckHealth(ctx context.Context, req *servicehealth.CheckRequest) (*servicehealth.HealthResponse, error)
	GetServiceInfo(ctx context.Context) map[string]interface{}
}

// Handler 健康检查 HTTP 处理器
type Handler struct {
	service Service
}

// New 创建健康检查处理器
func New(healthService Service) *Handler {
	return &Handler{
		service: healthService,
	}
}

// ServeHTTP 满足 http.Handler，但通常通过路由注册具体方法。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// SetupRoutes 注册路由
func (h *Handler) SetupRoutes(router interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request), ...string)
}) {
	router.HandleFunc("/api/health", h.CheckHealth, http.MethodGet)
	router.HandleFunc("/api/health/info", h.GetServiceInfo, http.MethodGet)
}

// CheckHealth 健康检查
func (h *Handler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.CheckHealth(r.Context(), nil)
	if err != nil {
		respond.Internal(w, r, "health check failed", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, resp)
}

// GetServiceInfo 获取服务信息
func (h *Handler) GetServiceInfo(w http.ResponseWriter, r *http.Request) {
	info := h.service.GetServiceInfo(r.Context())
	respond.JSON(w, http.StatusOK, r, info)
}
