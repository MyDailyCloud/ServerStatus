package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"serverstatus-monitor/internal/exportclean/application"
)

// NewHTTPHandler 返回一个最小的 HTTP 处理器，复用已有 Controller。
// 路由：POST /api/export/tasks
func NewHTTPHandler(ctrl *Controller) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/export/tasks" {
			http.NotFound(w, r)
			return
		}

		ct := r.Header.Get("Content-Type")
		if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "unsupported_media_type",
				"message": "Content-Type must be application/json",
			})
			return
		}

		var dto SubmitTaskDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": "invalid json",
			})
			return
		}

		err := ctrl.HandleSubmit(context.Background(), &dto)
		if err != nil {
			code := http.StatusInternalServerError
			switch err {
			case application.ErrInvalidFormat, application.ErrInvalidType, application.ErrInvalidLimit, application.ErrMissingID:
				code = http.StatusBadRequest
			default:
				// 检查其他校验错误信息
				if strings.Contains(err.Error(), "project_key is required") ||
					strings.Contains(err.Error(), "project_key too long") ||
					strings.Contains(err.Error(), "invalid characters") ||
					strings.Contains(err.Error(), "filter key") ||
					strings.Contains(err.Error(), "filter value") {
					code = http.StatusBadRequest
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_request",
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})
}
