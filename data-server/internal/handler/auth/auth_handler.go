package authhandler

import (
	"encoding/json"
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/internal/handler/respond"
	"github.com/kanshan/ServerStatus/data-server/internal/service"
)

// Handler 认证处理器
type Handler struct {
	service service.AuthService
}

// New 创建认证处理器
func New(authService service.AuthService) *Handler {
	return &Handler{
		service: authService,
	}
}

// SetupRoutes 注册路由
func (h *Handler) SetupRoutes(router interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request), ...string)
}) {
	router.HandleFunc("/api/auth/validate", h.ValidateAccessKey, http.MethodPost)
	router.HandleFunc("/api/auth/generate", h.GenerateAccessKey, http.MethodPost)
}

type validateRequest struct {
	AccessKey  string `json:"access_key"`
	ProjectKey string `json:"project_key"`
}

type validateResponse struct {
	Valid      bool   `json:"valid"`
	ProjectKey string `json:"project_key,omitempty"`
	Message    string `json:"message,omitempty"`
}

// ValidateAccessKey 验证访问密钥
func (h *Handler) ValidateAccessKey(w http.ResponseWriter, r *http.Request) {
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request", err)
		return
	}
	projectKey, err := h.service.ValidateAccessKey(r.Context(), req.AccessKey)
	if err != nil {
		respond.JSON(w, http.StatusUnauthorized, r, validateResponse{
			Valid:   false,
			Message: "invalid access key",
		})
		return
	}
	if req.ProjectKey != "" && projectKey != "*" && projectKey != req.ProjectKey {
		respond.JSON(w, http.StatusForbidden, r, validateResponse{
			Valid:   false,
			Message: "access key not allowed for project",
		})
		return
	}
	respond.JSON(w, http.StatusOK, r, validateResponse{
		Valid:      true,
		ProjectKey: projectKey,
		Message:    "ok",
	})
}

type generateRequest struct {
	ServerKey  string `json:"server_key"`
	ProjectKey string `json:"project_key"`
}

type generateResponse struct {
	AccessKey string `json:"access_key"`
}

// GenerateAccessKey 生成访问密钥
func (h *Handler) GenerateAccessKey(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request", err)
		return
	}
	accessKey, err := h.service.GenerateAccessKey(r.Context(), req.ServerKey, req.ProjectKey)
	if err != nil {
		respond.Internal(w, r, "failed to generate access key", err)
		return
	}
	respond.JSON(w, http.StatusOK, r, generateResponse{AccessKey: accessKey})
}
