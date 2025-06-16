package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ezhttp/ezhttp/internal/config"
	"github.com/ezhttp/ezhttp/internal/logger"
)

type Handler struct {
	targetURL *url.URL
	transport *http.Transport
	config    *config.DataConfigProxy
}

// Creates a new proxy handler
func NewHandler(cfg *config.DataConfigProxy) (*Handler, error) {
	if cfg.TargetURL == "" {
		return nil, fmt.Errorf("proxy target URL is required")
	}

	targetURL, err := url.Parse(cfg.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	h := &Handler{
		targetURL: targetURL,
		config:    cfg,
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
	if h.config.AllowInsecureOriginTLS {
		// Skip certificate verification (dangerous, but needed for some legacy systems)
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}
		logger.Warn("Proxy configured to skip TLS verification for origin")
	} else if h.config.RelaxedOriginTLS {
		// Use default cipher suites for compatibility
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		logger.Info("Proxy using relaxed TLS settings for origin")
	} else {
		// Strong TLS configuration
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			},
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
			},
		}
		logger.Info("Proxy using strong TLS settings for origin")
	}

	return transport
}

// Handles proxy requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create new request for backend
	targetURL := *h.targetURL
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		logger.Error("Failed to create proxy request", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	copyHeaders(proxyReq.Header, r.Header)

	// Remove hop-by-hop headers
	removeHopByHopHeaders(proxyReq.Header)

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

	// Send request to backend
	resp, err := h.transport.RoundTrip(proxyReq)
	if err != nil {
		logger.Error("Backend request failed", "error", err, "url", targetURL.String())
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	copyHeaders(w.Header(), resp.Header)
	removeHopByHopHeaders(w.Header())

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("Failed to copy response body", "error", err)
	}
}

// Copies headers from source to destination
func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
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
