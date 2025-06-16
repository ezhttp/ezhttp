package main

import (
	"flag"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ezhttp/ezhttp/internal/config"
	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/middleware"
	"github.com/ezhttp/ezhttp/internal/ratelimit"
	"github.com/ezhttp/ezhttp/internal/security"
	"github.com/ezhttp/ezhttp/internal/server"
	tlsconfig "github.com/ezhttp/ezhttp/internal/tls"
	"github.com/ezhttp/ezhttp/internal/version"

	"github.com/tdewolff/minify/v2"
	// TODO: Had issues with the other minifiers but will revisit
	//"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	// "github.com/tdewolff/minify/v2/js"
	// "github.com/tdewolff/minify/v2/json"
)

// ROADMAP
// TODO: Report URI Node.js / Golang
// TODO: LOGGING
// TODO: Check for index.html changes (datetime etc)
// TODO: Example favicon files mismatched
// TODO: Example CDN Host (img, script, css)
// TODO: Config HTTP Timeout
// TODO: Config Minification Settings
// TODO: Minification for CSS/JS (had issues)

var cfg config.DataConfig

// Internal
var compiledCsp string = ""
var cachedIndexString []string = []string{}
var minifier *minify.M

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
	cfg = config.ConfigLoad()

	// Cache Generated Index and CSP
	cachedIndexString, _ = server.LoadIndexCache()
	compiledCsp = cfg.Csp.Compile()

	// Set Up Minification
	minifier = minify.New()
	// TODO: Had issues with the other minifiers but will revisit
	//minifier.AddFunc("text/css", css.Minify)
	minifier.Add("text/html", &html.Minifier{
		KeepConditionalComments: false,
		KeepDocumentTags:        true,
		KeepDefaultAttrVals:     false,
		KeepWhitespace:          false,
		KeepEndTags:             true,
		KeepQuotes:              true,
	})
	//minifier.AddFunc("image/svg+xml", svg.Minify)
	// minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	// minifier.AddFuncRegexp(regexp.MustCompile("^application/json$"), json.Minify)

	// Create secure file server with custom FileSystem
	publicDir, err := filepath.Abs("./public")
	if err != nil {
		logger.Fatal("Failed to resolve public directory")
	}

	// Use a custom FileSystem that prevents directory listings and symlink attacks
	secureFS := &security.SecureFileSystem{Fs: http.Dir(publicDir)}
	httpfs := http.FileServer(secureFS)
	// //http.Handle("/static/", http.StripPrefix("/public/", httpfs))
	// // strings.TrimRight("/statics/", "/")
	// http.Handle("/", mwNonce(httpfs))

	// Create handler chain
	var handler http.Handler = server.MwNonce(httpfs, compiledCsp, cachedIndexString, minifier, cfg.Banner)

	// Apply security headers middleware
	handler = middleware.SecurityHeadersMiddleware(handler)

	// Apply rate limiting if enabled
	if cfg.RateLimit.Enabled {
		cleanupInterval := server.ParseCleanupInterval(cfg.RateLimit.CleanupInterval)
		limiter := ratelimit.NewLimiter(
			cfg.RateLimit.RequestsPerMinute,
			cfg.RateLimit.BurstSize,
			cleanupInterval,
		)
		handler = middleware.RateLimitMiddleware(limiter)(handler)
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

		logger.Info("Starting HTTPS server",
			"address", cfg.ListenAddr,
			"port", cfg.ListenPort,
			"cert", cfg.TLS.CertFile,
			"key", cfg.TLS.KeyFile)

		ln, err := net.Listen(network, httpServer.Addr)
		if err != nil {
			logger.Fatal("Failed to listen", "error", err)
		}
		err = httpServer.ServeTLS(ln, cfg.TLS.CertFile, cfg.TLS.KeyFile)
	} else {
		logger.Info("Starting HTTP server", "address", cfg.ListenAddr, "port", cfg.ListenPort)
		ln, err := net.Listen(network, httpServer.Addr)
		if err != nil {
			logger.Fatal("Failed to listen", "error", err)
		}
		err = httpServer.Serve(ln)
	}

	if err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
