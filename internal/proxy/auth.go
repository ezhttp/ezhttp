package proxy

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
)

// Creates an authentication middleware for the proxy with IP blocking support
func AuthMiddleware(authToken string, limiter *ProxyLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no auth token is configured, allow all requests
			if authToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			ip := ratelimit.ExtractIP(r.RemoteAddr)

			// Check if IP is blocked
			if limiter != nil && limiter.IsBlocked(ip) {
				logger.Warn("Blocked IP attempted access", "ip", ip)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check X-Proxy-Password header first (recommended)
			proxyPassword := r.Header.Get("X-Proxy-Password")
			if proxyPassword != "" {
				if subtle.ConstantTimeCompare([]byte(proxyPassword), []byte(authToken)) == 1 {
					// Authentication successful
					if limiter != nil {
						limiter.ResetAuthFailures(ip)
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check Authorization Bearer header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				const bearerPrefix = "Bearer "
				if strings.HasPrefix(authHeader, bearerPrefix) {
					token := authHeader[len(bearerPrefix):]
					if subtle.ConstantTimeCompare([]byte(token), []byte(authToken)) == 1 {
						// Authentication successful
						if limiter != nil {
							limiter.ResetAuthFailures(ip)
						}
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// Authentication failed
			if limiter != nil {
				limiter.RecordAuthFailure(ip)
			}
			w.Header().Set("WWW-Authenticate", `Bearer realm="proxy"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// HealthCheckMiddleware bypasses authentication for health check endpoint
func HealthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check endpoint - no auth required
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
