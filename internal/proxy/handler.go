package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ezhttp/ezhttp/internal/config"
	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
	tlsconfig "github.com/ezhttp/ezhttp/internal/tls"
)

type Handler struct {
	originBaseURL *url.URL
	transport     *http.Transport
	config        *config.DataConfigProxy
	debugMode     bool
}

// Creates a new proxy handler
func NewHandler(cfg *config.DataConfigProxy) (*Handler, error) {
	if cfg.OriginBaseURL == "" {
		return nil, fmt.Errorf("proxy origin base URL is required")
	}

	originBaseURL, err := url.Parse(cfg.OriginBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid origin base URL: %w", err)
	}

	h := &Handler{
		originBaseURL: originBaseURL,
		config:        cfg,
		debugMode:     cfg.DebugMode,
	}

	// Create transport with security settings
	h.transport = h.createTransport()

	return h, nil
}

// Creates HTTP transport with appropriate security settings
func (h *Handler) createTransport() *http.Transport {
	// Parse idle connection timeout
	idleConnTimeout := 90 * time.Second
	if h.config.IdleConnTimeout != "" {
		if d, err := time.ParseDuration(h.config.IdleConnTimeout); err == nil {
			idleConnTimeout = d
		}
	}

	transport := &http.Transport{
		MaxIdleConns:        h.config.MaxIdleConns,
		IdleConnTimeout:     idleConnTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  false,
	}

	// Configure TLS based on settings
	var securityLevel tlsconfig.SecurityLevel
	if h.config.AllowInsecureOriginTLS {
		securityLevel = tlsconfig.SecurityLevelInsecure
	} else if h.config.RelaxedOriginTLS {
		securityLevel = tlsconfig.SecurityLevelRelaxed
	} else {
		securityLevel = tlsconfig.SecurityLevelStrong
	}
	transport.TLSClientConfig = tlsconfig.CreateClientTLSConfig(securityLevel)

	return transport
}

// Handles proxy requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log the request
	clientIP := ratelimit.ExtractIP(r.RemoteAddr)

	// Debug logging for incoming request
	if h.debugMode {
		logger.Debug("Incoming proxy request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"ip", clientIP,
			"headers", fmt.Sprintf("%v", r.Header))
	}

	// Create new request for backend
	originURL := *h.originBaseURL
	originURL.Path = r.URL.Path
	originURL.RawQuery = r.URL.RawQuery

	proxyReq, err := http.NewRequest(r.Method, originURL.String(), r.Body)
	if err != nil {
		logger.Error("Failed to create proxy request", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Clone headers efficiently (available in Go 1.13+)
	proxyReq.Header = r.Header.Clone()

	// Remove hop-by-hop headers
	removeHopByHopHeaders(proxyReq.Header)

	// Remove sensitive headers before forwarding
	removeSensitiveRequestHeaders(proxyReq.Header)

	// Add X-Forwarded headers
	if clientIP := r.RemoteAddr; clientIP != "" {
		if prior, ok := proxyReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
	}
	proxyReq.Header.Set("X-Forwarded-Proto", "http")
	if r.TLS != nil {
		proxyReq.Header.Set("X-Forwarded-Proto", "https")
	}
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)
	proxyReq.Header.Set("X-Real-IP", r.RemoteAddr)

	// Debug logging for outgoing request
	if h.debugMode {
		logger.Debug("Outgoing backend request",
			"method", proxyReq.Method,
			"url", originURL.String(),
			"headers", fmt.Sprintf("%v", proxyReq.Header))
	}

	// Send request to backend
	resp, err := h.transport.RoundTrip(proxyReq)
	if err != nil {
		logger.Error("Backend request failed", "error", err, "url", originURL.String())

		// Check if it's a timeout error
		if err, ok := err.(net.Error); ok && err.Timeout() {
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		}

		// Check for context deadline exceeded
		if strings.Contains(err.Error(), "context deadline exceeded") {
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		}

		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Debug logging for backend response
	if h.debugMode {
		logger.Debug("Backend response",
			"status", resp.StatusCode,
			"headers", fmt.Sprintf("%v", resp.Header))
	}

	// Copy response headers efficiently
	responseHeaders := w.Header()
	for k, vv := range resp.Header {
		responseHeaders[k] = vv
	}
	removeHopByHopHeaders(responseHeaders)
	removeSensitiveResponseHeaders(responseHeaders)

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if h.debugMode {
		// In debug mode, capture and log response body (limited to first 1KB)
		var buf [1024]byte
		n, _ := resp.Body.Read(buf[:])
		if n > 0 {
			logger.Debug("Response body preview",
				"size", n,
				"content", string(buf[:n]))
			// Write what we read
			w.Write(buf[:n])
		}
		// Copy the rest
		_, err = io.Copy(w, resp.Body)
	} else {
		// Normal mode - just copy
		_, err = io.Copy(w, resp.Body)
	}
	if err != nil {
		logger.Error("Failed to copy response body", "error", err)
	}

	// Log successful request
	logger.Info("Proxy request",
		"method", r.Method,
		"path", r.URL.Path,
		"target", originURL.String(),
		"status", resp.StatusCode,
		"ip", clientIP)
}

// Removes headers that shouldn't be forwarded
func removeHopByHopHeaders(h http.Header) {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, header := range hopByHopHeaders {
		h.Del(header)
	}
}

// Removes sensitive headers before forwarding
func removeSensitiveRequestHeaders(h http.Header) {
	sensitiveHeaders := []string{
		"Authorization",
		"Cookie",
		"Set-Cookie",
		"X-Proxy-Password",
		"X-Api-Key",
		"X-Auth-Token",
		"X-Access-Token",
		"X-Secret-Token",
		"Api-Key",
		"Access-Token",
		"Auth-Token",
	}

	for _, header := range sensitiveHeaders {
		h.Del(header)
	}
}

// Removes headers that reveal server info
func removeSensitiveResponseHeaders(h http.Header) {
	h.Del("Server")
	h.Del("X-Powered-By")
}
