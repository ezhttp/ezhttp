package main

import (
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/ezhttp/ezhttp/internal/config"
	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/middleware"
	"github.com/ezhttp/ezhttp/internal/proxy"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
	"github.com/ezhttp/ezhttp/internal/server"
	tlsconfig "github.com/ezhttp/ezhttp/internal/tls"
	"github.com/ezhttp/ezhttp/internal/version"
)

func main() {
	// Flag Check
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()
	if *showVersion {
		version.PrintVersion()
		return
	}

	version.PrintVersionShort()

	// Load Config
	cfg := config.ConfigLoad()

	// Validate proxy configuration
	if cfg.Proxy.TargetURL == "" {
		logger.Fatal("Proxy target URL is required. Set proxy.target_url in config or PROXY_TARGET environment variable")
	}

	// Create proxy handler
	proxyHandler, err := proxy.NewHandler(&cfg.Proxy)
	if err != nil {
		logger.Fatal("Failed to create proxy handler", "error", err)
	}

	// Build middleware chain (order matters!)
	var handler http.Handler = proxyHandler

	// Add security headers (applies to all responses)
	handler = middleware.SecurityHeadersMiddleware(handler)

	// Create proxy limiter with IP blocking if auth is enabled
	var proxyLimiter *proxy.ProxyLimiter
	if cfg.Proxy.AuthToken != "" {
		// Parse block duration
		blockDuration := 15 * time.Minute
		if cfg.Proxy.BlockDuration != "" {
			if d, err := time.ParseDuration(cfg.Proxy.BlockDuration); err == nil {
				blockDuration = d
			}
		}

		cleanupInterval := server.ParseCleanupInterval(cfg.RateLimit.CleanupInterval)
		proxyLimiter = proxy.NewProxyLimiter(
			cfg.RateLimit.RequestsPerMinute,
			cfg.RateLimit.BurstSize,
			cfg.Proxy.MaxAuthAttempts,
			cleanupInterval,
			blockDuration,
		)

		// Add authentication middleware with IP blocking
		handler = proxy.AuthMiddleware(cfg.Proxy.AuthToken, proxyLimiter)(handler)
		logger.Info("Proxy authentication enabled",
			"max_auth_attempts", cfg.Proxy.MaxAuthAttempts,
			"block_duration", blockDuration.String())
	}

	// Add health check endpoint (bypasses auth)
	handler = proxy.HealthCheckMiddleware(handler)

	// Add request validation (size limiting, host validation, etc.)
	handler = proxy.RequestValidationMiddleware(cfg.Proxy.AllowedHost, cfg.Proxy.MaxRequestSize)(handler)

	// Apply rate limiting if enabled
	if cfg.RateLimit.Enabled {
		if proxyLimiter != nil {
			// Use proxy limiter for rate limiting
			handler = middleware.ProxyRateLimitMiddleware(proxyLimiter)(handler)
		} else {
			// Use basic rate limiter
			cleanupInterval := server.ParseCleanupInterval(cfg.RateLimit.CleanupInterval)
			limiter := ratelimit.NewLimiter(
				cfg.RateLimit.RequestsPerMinute,
				cfg.RateLimit.BurstSize,
				cleanupInterval,
			)
			handler = middleware.RateLimitMiddleware(limiter)(handler)
		}
		logger.Info("Rate limiting enabled",
			"requests_per_minute", cfg.RateLimit.RequestsPerMinute,
			"burst_size", cfg.RateLimit.BurstSize)
	}

	// Configure server
	httpServer := &http.Server{
		Addr:              cfg.ListenAddr + ":" + cfg.ListenPort,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           handler,
	}

	// Always use IPv4-only mode
	network := "tcp4"

	// Start server with or without TLS
	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		// Configure TLS
		httpServer.TLSConfig = tlsconfig.CreateServerTLSConfig()

		logger.Info("Starting HTTPS proxy server",
			"address", cfg.ListenAddr,
			"port", cfg.ListenPort,
			"target", cfg.Proxy.TargetURL,
			"cert", cfg.TLS.CertFile,
			"key", cfg.TLS.KeyFile)

		ln, err := net.Listen(network, httpServer.Addr)
		if err != nil {
			logger.Fatal("Failed to listen", "error", err)
		}
		err = httpServer.ServeTLS(ln, cfg.TLS.CertFile, cfg.TLS.KeyFile)
	} else {
		logger.Info("Starting HTTP proxy server",
			"address", cfg.ListenAddr,
			"port", cfg.ListenPort,
			"target", cfg.Proxy.TargetURL)
		ln, err := net.Listen(network, httpServer.Addr)
		if err != nil {
			logger.Fatal("Failed to listen", "error", err)
		}
		err = httpServer.Serve(ln)
	}

	if err != nil {
		logger.Fatal("Proxy server failed to start", "error", err)
	}
}
