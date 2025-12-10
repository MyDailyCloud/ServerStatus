package middleware

import (
	"net/http"
	"strings"
)

// RequireJSON 确保写操作使用 application/json，避免内容类型不一致。
func RequireJSON() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 仅对可能有请求体的方法校验
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				ct := r.Header.Get("Content-Type")
				if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = w.Write([]byte(`{"error":"unsupported_media_type","message":"Content-Type must be application/json"}`))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
