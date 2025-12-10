package exporthandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/internal/handler/respond"
	"github.com/kanshan/ServerStatus/data-server/internal/service"
)

// Handler 导出相关 HTTP 处理器
type Handler struct {
	service service.ExportService
}

// New 创建导出处理器
func New(exportService service.ExportService) *Handler {
	return &Handler{
		service: exportService,
	}
}

// SetupRoutes 注册路由
func (h *Handler) SetupRoutes(router interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request), ...string)
}) {
	router.HandleFunc("/api/export/formats", h.GetFormats, http.MethodGet)
	router.HandleFunc("/api/export/servers", h.ExportServers, http.MethodPost)
	router.HandleFunc("/api/export/history", h.ExportHistory, http.MethodPost)
	router.HandleFunc("/api/export/user-resources", h.ExportUserResources, http.MethodPost)
	router.HandleFunc("/api/export/preview", h.GetPreview, http.MethodPost)
}

// GetFormats 返回支持的导出格式
func (h *Handler) GetFormats(w http.ResponseWriter, r *http.Request) {
	formats := h.service.GetSupportedFormats()
	respond.JSON(w, http.StatusOK, r, map[string]interface{}{
		"formats": formats,
	})
}

// ExportServers 导出服务器数据
func (h *Handler) ExportServers(w http.ResponseWriter, r *http.Request) {
	h.export(w, r, h.service.ExportServers)
}

// ExportHistory 导出历史数据
func (h *Handler) ExportHistory(w http.ResponseWriter, r *http.Request) {
	h.export(w, r, h.service.ExportHistory)
}

// ExportUserResources 导出用户资源数据
func (h *Handler) ExportUserResources(w http.ResponseWriter, r *http.Request) {
	h.export(w, r, h.service.ExportUserResources)
}

// GetPreview 获取导出预览
func (h *Handler) GetPreview(w http.ResponseWriter, r *http.Request) {
	var req service.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request", err)
		return
	}
	if err := h.service.ValidateExportRequest(&req); err != nil {
		respond.BadRequest(w, r, "invalid export request", err)
		return
	}
	preview, err := h.service.GetExportPreview(r.Context(), &req)
	if err != nil {
		respond.Internal(w, r, "failed to get export preview", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, preview)
}

func (h *Handler) export(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *service.ExportRequest) (*service.ExportResult, error)) {
	var req service.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request", err)
		return
	}
	if err := h.service.ValidateExportRequest(&req); err != nil {
		respond.BadRequest(w, r, "invalid export request", err)
		return
	}

	// 调用具体导出函数
	result, err := fn(r.Context(), &req)
	if err != nil {
		respond.Internal(w, r, "export failed", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, result)
}
