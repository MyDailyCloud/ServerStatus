package websockethandler

import (
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/internal/handler/respond"
	"github.com/kanshan/ServerStatus/data-server/internal/service"
)

// Handler WebSocket 相关 HTTP 处理器（只提供查询，不负责升级）。
type Handler struct {
	service service.WebSocketService
}

// New 创建 WebSocket 处理器
func New(wsService service.WebSocketService) *Handler {
	return &Handler{
		service: wsService,
	}
}

// SetupRoutes 注册路由
func (h *Handler) SetupRoutes(router interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request), ...string)
}) {
	router.HandleFunc("/api/ws/stats", h.GetStats, http.MethodGet)
	router.HandleFunc("/api/ws/connections", h.GetConnections, http.MethodGet)
}

// GetStats 返回连接统计
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.service.GetConnectionStats()
	respond.JSON(w, http.StatusOK, r, stats)
}

// GetConnections 返回连接列表（基础信息）
func (h *Handler) GetConnections(w http.ResponseWriter, r *http.Request) {
	conns := h.service.GetConnections()
	respond.JSON(w, http.StatusOK, r, conns)
}
