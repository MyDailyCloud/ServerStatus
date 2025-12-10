package middleware

import (
	"net/http"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Recovery 捕获 panic，返回 500，防止异常冒泡。
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.WithField("panic", rec).Error("panic recovered")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
