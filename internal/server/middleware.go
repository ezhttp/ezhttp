package server

import (
	"io"
	"net/http"
	"time"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
)

// Creates a rate limiting middleware
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

// Parses a duration string for cleanup interval
func ParseCleanupInterval(interval string) time.Duration {
	if interval == "" {
		return 30 * time.Minute // Default to 30 minutes
	}

	d, err := time.ParseDuration(interval)
	if err != nil {
		logger.Warn("Invalid cleanup interval, using default", "interval", interval, "error", err)
		return 30 * time.Minute
	}

	return d
}
