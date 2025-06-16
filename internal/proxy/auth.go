package proxy

import (
	"net/http"
	"strings"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
)

// Creates an authentication middleware for the proxy
func AuthMiddleware(authToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no auth token is configured, allow all requests
			if authToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing authorization header", "ip", ratelimit.ExtractIP(r.RemoteAddr))
				w.Header().Set("WWW-Authenticate", `Bearer realm="proxy"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate Bearer token
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				logger.Warn("Invalid authorization header format", "ip", ratelimit.ExtractIP(r.RemoteAddr))
				w.Header().Set("WWW-Authenticate", `Bearer realm="proxy"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := authHeader[len(bearerPrefix):]
			if token != authToken {
				logger.Warn("Invalid authorization token", "ip", ratelimit.ExtractIP(r.RemoteAddr))
				w.Header().Set("WWW-Authenticate", `Bearer realm="proxy"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Authentication successful
			next.ServeHTTP(w, r)
		})
	}
}

// Adds security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Add HSTS header if using HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
