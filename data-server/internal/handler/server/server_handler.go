package serverhandler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kanshan/ServerStatus/data-server/internal/handler/respond"
	"github.com/kanshan/ServerStatus/data-server/internal/service"
)

// Handler 服务器相关 HTTP 处理器
type Handler struct {
	service service.ServerService
}

// New 创建服务器处理器
func New(serverService service.ServerService) *Handler {
	return &Handler{
		service: serverService,
	}
}

// SetupRoutes 注册路由
func (h *Handler) SetupRoutes(router interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request), ...string)
}) {
	router.HandleFunc("/api/servers", h.GetServers, http.MethodGet)
	router.HandleFunc("/api/servers/{session_id}", h.GetServer, http.MethodGet)
	router.HandleFunc("/api/projects/{project_key}/servers", h.GetServersByProject, http.MethodGet)
}

// GetServers 获取服务器列表
func (h *Handler) GetServers(w http.ResponseWriter, r *http.Request) {
	filter := &service.ServerFilter{
		ProjectKey: r.URL.Query().Get("project_key"),
		Hostname:   r.URL.Query().Get("hostname"),
		Status:     r.URL.Query().Get("status"),
		OS:         r.URL.Query().Get("os"),
	}
	servers, err := h.service.GetServers(r.Context(), filter)
	if err != nil {
		respond.Internal(w, r, "failed to get servers", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, servers)
}

// GetServer 获取单个服务器
func (h *Handler) GetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["session_id"]
	server, err := h.service.GetServer(r.Context(), sessionID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, r, "not_found", "server not found", err.Error())
		return
	}
	respond.JSON(w, http.StatusOK, r, server)
}

// GetServersByProject 按项目获取服务器
func (h *Handler) GetServersByProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectKey := vars["project_key"]
	pagination := parsePagination(r)
	servers, err := h.service.GetServersByProject(r.Context(), projectKey, pagination)
	if err != nil {
		respond.Internal(w, r, "failed to get project servers", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, servers)
}

func parsePagination(r *http.Request) *service.Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}
	return &service.Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}
}
