package middleware

import (
	"io"
	"net/http"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
)

// Creates a rate limiting middleware that works for both server and proxy
func RateLimitMiddleware(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP
			ip := ratelimit.ExtractIP(r.RemoteAddr)

			// Check rate limit
			if !limiter.Allow(ip) {
				logger.Warn("Rate limit exceeded", "ip", ip, "path", r.URL.Path)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("Retry-After", "60") // Suggest retry after 60 seconds
				w.WriteHeader(http.StatusTooManyRequests)
				io.WriteString(w, "Too Many Requests")
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// Creates a rate limiting middleware for proxy with IP blocking support
func ProxyRateLimitMiddleware(limiter interface{ Allow(string) bool }) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ratelimit.ExtractIP(r.RemoteAddr)
			if !limiter.Allow(ip) {
				logger.Warn("Rate limit exceeded", "ip", ip)
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
