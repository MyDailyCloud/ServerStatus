package respond

import (
	"encoding/json"
	"net/http"
)

const headerRequestID = "X-Request-ID"

// JSON 写入成功响应
func JSON(w http.ResponseWriter, status int, r *http.Request, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if rid := r.Header.Get(headerRequestID); rid != "" {
		w.Header().Set(headerRequestID, rid)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Error 写入统一错误响应
func Error(w http.ResponseWriter, status int, r *http.Request, code string, message string, detail string) {
	resp := map[string]interface{}{
		"error":   code,
		"message": message,
	}
	if detail != "" {
		resp["details"] = detail
	}
	if rid := r.Header.Get(headerRequestID); rid != "" {
		resp["req_id"] = rid
		w.Header().Set(headerRequestID, rid)
	}
	resp["path"] = r.URL.Path

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// BadRequest 400 错误响应
func BadRequest(w http.ResponseWriter, r *http.Request, message string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	Error(w, http.StatusBadRequest, r, "bad_request", message, detail)
}

// Internal 500 错误响应
func Internal(w http.ResponseWriter, r *http.Request, message string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	Error(w, http.StatusInternalServerError, r, "internal_error", message, detail)
}
