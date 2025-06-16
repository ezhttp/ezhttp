package proxy

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

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

			// Check for path traversal attempts
			if containsPathTraversal(r.URL.Path) {
				logger.Warn("Path traversal attempt", "path", r.URL.Path, "ip", ip)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			// Check request size from Content-Length header
			if r.ContentLength > 0 && r.ContentLength > maxRequestSize {
				logger.Warn("Request too large", "size", r.ContentLength, "max", maxRequestSize, "ip", ip)
				http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
				return
			}

			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

			next.ServeHTTP(w, r)
		})
	}
}

// Checks for common path traversal attempts using multiple validation methods
func containsPathTraversal(path string) bool {
	// First, decode any URL encoding multiple times to catch double-encoding
	decodedPath := path
	for i := 0; i < 3; i++ {
		decoded, err := url.QueryUnescape(decodedPath)
		if err != nil || decoded == decodedPath {
			break
		}
		decodedPath = decoded
	}

	// Clean the path
	cleanPath := filepath.Clean("/" + decodedPath)

	// Check if the cleaned path tries to escape the root
	if !strings.HasPrefix(cleanPath, "/") {
		return true
	}

	// Use filepath.Rel to check if path stays within bounds
	// Convert to a simulated absolute path for checking
	basePath := "/tmp/base"
	targetPath := filepath.Join(basePath, cleanPath)
	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil || strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
		return true
	}

	// Additional pattern matching for various encoding attempts
	traversalPatterns := []string{
		"..",
		"..\\",
		"../",
		".%2e",
		"%2e.",
		"%2e%2e",
		"%252e",
		"..%5c",
		"%5c..",
		"..%255c",
		"%c0%ae",
		"%c1%9c",
		"\\x2e\\x2e",
		"0x2e0x2e",
	}

	lowerPath := strings.ToLower(decodedPath)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}
