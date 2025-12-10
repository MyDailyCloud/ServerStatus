package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// statusRecorder 记录响应状态码以便日志输出。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging 简单请求日志中间件，输出 method/path/status/duration。
func Logging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			reqID := r.Header.Get(headerRequestID)
			clientIP := clientIP(r)

			next.ServeHTTP(rec, r)

			log.WithFields(map[string]interface{}{
				"method":   r.Method,
				"path":     r.URL.Path,
				"status":   rec.status,
				"duration": time.Since(start).String(),
				"req_id":   reqID,
				"ip":       clientIP,
			}).Info("http request")
		})
	}
}

func clientIP(r *http.Request) string {
	// 优先使用 X-Forwarded-For 第一个值，其次 RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 可能是逗号分隔
		for _, part := range strings.Split(xff, ",") {
			if ip := strings.TrimSpace(part); ip != "" {
				return ip
			}
		}
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	// RemoteAddr 可能包含端口
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
