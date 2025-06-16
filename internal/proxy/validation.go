package proxy

import (
	"io"
	"net/http"
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

// Checks for common path traversal attempts
func containsPathTraversal(path string) bool {
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
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// SizeLimitMiddleware creates a middleware specifically for request size limiting
func SizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check Content-Length header first
			if r.ContentLength > maxSize {
				logger.Warn("Request size exceeds limit",
					"content_length", r.ContentLength,
					"max_size", maxSize,
					"ip", ratelimit.ExtractIP(r.RemoteAddr))

				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				io.WriteString(w, "Request Entity Too Large")
				return
			}

			// Wrap body with size limiter
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			// Create a custom response writer to catch MaxBytesReader errors
			wrapped := &sizeErrorResponseWriter{
				ResponseWriter: w,
				maxSize:        maxSize,
				request:        r,
			}

			next.ServeHTTP(wrapped, r)
		})
	}
}

// Wraps http.ResponseWriter to catch size limit errors
type sizeErrorResponseWriter struct {
	http.ResponseWriter
	maxSize int64
	request *http.Request
	written bool
}

func (w *sizeErrorResponseWriter) Write(b []byte) (int, error) {
	if w.written {
		return 0, nil
	}
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		// Check if it's a MaxBytesReader error
		if strings.Contains(err.Error(), "http: request body too large") {
			if !w.written {
				w.written = true
				logger.Warn("Request body too large during read",
					"max_size", w.maxSize,
					"ip", ratelimit.ExtractIP(w.request.RemoteAddr))
			}
		}
	}
	return n, err
}

func (w *sizeErrorResponseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.written = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}
