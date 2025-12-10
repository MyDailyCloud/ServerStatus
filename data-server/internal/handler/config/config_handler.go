package confighandler

import (
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/internal/handler/respond"
	"github.com/kanshan/ServerStatus/data-server/internal/service"
)

// Handler 配置相关 HTTP 处理器
type Handler struct {
	service service.ConfigService
}

// New 创建配置处理器
func New(configService service.ConfigService) *Handler {
	return &Handler{
		service: configService,
	}
}

// SetupRoutes 注册路由
func (h *Handler) SetupRoutes(router interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request), ...string)
}) {
	router.HandleFunc("/api/config", h.GetConfig, http.MethodGet)
	router.HandleFunc("/api/config/reload", h.ReloadConfig, http.MethodPost)
}

// GetConfig 返回当前配置
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.service.GetConfig()
	respond.JSON(w, http.StatusOK, r, cfg)
}

// ReloadConfig 触发配置重载
func (h *Handler) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	if err := h.service.ReloadConfig(); err != nil {
		respond.Internal(w, r, "failed to reload config", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, map[string]string{"status": "reloaded"})
}
