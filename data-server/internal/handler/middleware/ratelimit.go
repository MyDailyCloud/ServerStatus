package middleware

import (
	"net/http"
	"sync"
	"time"
)

// 简易基于内存的固定窗口限流，按 remote IP 计数。
// 目标是提供可选保护，不追求分布式/精确。
type RateLimiter struct {
	requests   map[string]int
	mu         sync.Mutex
	limit      int
	window     time.Duration
	lastWindow time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:   make(map[string]int),
		limit:      limit,
		window:     window,
		lastWindow: time.Now(),
	}
}

// Middleware 返回中间件，超限返回 429。
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			rl.mu.Lock()
			now := time.Now()
			if now.Sub(rl.lastWindow) > rl.window {
				rl.requests = make(map[string]int)
				rl.lastWindow = now
			}
			rl.requests[ip]++
			count := rl.requests[ip]
			rl.mu.Unlock()

			if count > rl.limit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate_limited","message":"too many requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
