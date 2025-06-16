package middleware

import (
	"net/http"
)

// Adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		// Add HSTS header if using HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Remove server identification headers
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")

		next.ServeHTTP(w, r)
	})
}

// Adds Content-Security-Policy header with nonce support
func CSPMiddleware(csp string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if csp != "" {
				w.Header().Set("Content-Security-Policy", csp)
			}
			next.ServeHTTP(w, r)
		})
	}
}
