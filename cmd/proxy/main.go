package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/ezhttp/ezhttp/internal/config"
	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/proxy"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
	"github.com/ezhttp/ezhttp/internal/server"
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

	// Build middleware chain
	var handler http.Handler = proxyHandler

	// Add security headers
	handler = proxy.SecurityHeadersMiddleware(handler)

	// Add authentication if configured
	if cfg.Proxy.AuthToken != "" {
		handler = proxy.AuthMiddleware(cfg.Proxy.AuthToken)(handler)
		logger.Info("Proxy authentication enabled")
	}

	// Apply rate limiting if enabled
	if cfg.RateLimit.Enabled {
		cleanupInterval := server.ParseCleanupInterval(cfg.RateLimit.CleanupInterval)
		limiter := ratelimit.NewLimiter(
			cfg.RateLimit.RequestsPerMinute,
			cfg.RateLimit.BurstSize,
			cleanupInterval,
		)
		handler = server.RateLimitMiddleware(limiter)(handler)
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
		Handler:           handler,
	}

	// Start server with or without TLS
	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		// Configure TLS
		httpServer.TLSConfig = server.CreateTLSConfig()

		logger.Info("Starting HTTPS proxy server",
			"address", cfg.ListenAddr,
			"port", cfg.ListenPort,
			"target", cfg.Proxy.TargetURL,
			"cert", cfg.TLS.CertFile,
			"key", cfg.TLS.KeyFile)

		err = httpServer.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	} else {
		logger.Info("Starting HTTP proxy server",
			"address", cfg.ListenAddr,
			"port", cfg.ListenPort,
			"target", cfg.Proxy.TargetURL)
		err = httpServer.ListenAndServe()
	}

	if err != nil {
		logger.Fatal("Proxy server failed to start", "error", err)
	}
}
