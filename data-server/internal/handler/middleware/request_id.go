package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

// RequestID 确保每个请求都有唯一 ID，便于日志/追踪。
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(headerRequestID)
			if reqID == "" {
				reqID = uuid.NewString()
			}
			w.Header().Set(headerRequestID, reqID)
			r.Header.Set(headerRequestID, reqID)
			next.ServeHTTP(w, r)
		})
	}
}
