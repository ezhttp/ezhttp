package proxy

import (
	"net/http"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
)

// AllowedMethods defines the HTTP methods allowed through the proxy
var AllowedMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
}

// Validates incoming requests
func RequestValidationMiddleware(allowedHost string, maxRequestSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ratelimit.ExtractIP(r.RemoteAddr)

			// Validate HTTP method
			if !AllowedMethods[r.Method] {
				logger.Warn("Invalid HTTP method", "method", r.Method, "ip", ip)
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			// Validate host header
			if allowedHost != "" && r.Host != allowedHost {
				logger.Warn("Invalid host header", "host", r.Host, "expected", allowedHost, "ip", ip)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
