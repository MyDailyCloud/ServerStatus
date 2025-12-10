package middleware

import (
	"net/http"
	"strings"
)

// CORS 简易 CORS 中间件，支持配置允许的源/方法/头。
func CORS(origins, methods, headers []string) func(http.Handler) http.Handler {
	allowOrigin := strings.Join(origins, ", ")
	allowMethods := strings.Join(methods, ", ")
	allowHeaders := strings.Join(headers, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
