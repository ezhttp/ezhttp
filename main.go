package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ezhttp/ezhttp/internal/config"
	"github.com/ezhttp/ezhttp/internal/security"
	"github.com/ezhttp/ezhttp/internal/server"
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
		log.Fatal("Failed to resolve public directory:", err)
	}

	// Use a custom FileSystem that prevents directory listings and symlink attacks
	secureFS := &security.SecureFileSystem{Fs: http.Dir(publicDir)}
	httpfs := http.FileServer(secureFS)
	// //http.Handle("/static/", http.StripPrefix("/public/", httpfs))
	// // strings.TrimRight("/statics/", "/")
	// http.Handle("/", mwNonce(httpfs))

	log.Printf("[SERVER] Listening on %s:%s...\n", cfg.ListenAddr, cfg.ListenPort)
	httpServer := &http.Server{
		Addr:              cfg.ListenAddr + ":" + cfg.ListenPort,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		//TLSConfig:         tlsConfig,
		Handler: server.MwNonce(httpfs, compiledCsp, cachedIndexString, minifier, cfg.Banner),
	}
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
