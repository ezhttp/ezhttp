package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"dario.cat/mergo"
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

// Build Information
const BuildVersion string = "0.0.2"

var BuildDate string
var BuildGoVersion string
var BuildGitHash string

var config DataConfig

// Internal
var compiledCsp string = ""
var cachedIndexString []string = []string{}
var minifier *minify.M

// =========================
// ======== Config =========
// =========================

type DataConfig struct {
	Version          int           `json:"version"`
	ListenAddr       string        `json:"listen_addr"`
	ListenPort       string        `json:"listen_port"`
	NoncePlaceholder string        `json:"nonce_placeholder"`
	Banner           []string      `json:"banner"`
	Csp              DataConfigCsp `json:"csp"`
}

type DataConfigCsp struct {
	DefaultSrc             string   `json:"default-src"`
	BaseUri                string   `json:"base-uri"`
	ConnectSrc             []string `json:"connect-src"`
	FontSrc                []string `json:"font-src"`
	FormAction             []string `json:"form-action"`
	FrameAncestors         []string `json:"frame-ancestors"`
	FrameSrc               []string `json:"frame-src"`
	ImgSrc                 []string `json:"img-src"`
	ManifestSrc            []string `json:"manifest-src"`
	MediaSrc               []string `json:"media-src"`
	ObjectSrc              []string `json:"object-src"`
	RequireTrustedTypesFor []string `json:"require-trusted-types-for"`
	ScriptSrc              []string `json:"script-src"`
	StyleSrc               []string `json:"style-src"`
}

func DefaultConfigCsp() DataConfigCsp {
	// sandbox
	// TODO: Setup
	//"report-uri https://csp.example.com;",
	// child-src
	//"report-to https://csp.example.com;",
	//"prefetch-src 'none'",
	// navigate-to
	// upgrade-insecure-requests
	// block-all-mixed-content
	// Do not use. Removed
	//"plugin-types 'none';",
	return DataConfigCsp{
		// TODO: Move to 'none'. Angular breaks for some reason
		DefaultSrc: "'self'",
		BaseUri:    "'self'",
		ConnectSrc: []string{
			"'self'",
			"https://fonts.gstatic.com",
		},
		FontSrc: []string{
			"'self'",
			"fonts.gstatic.com",
		},
		FormAction: []string{
			"'self'",
		},
		FrameAncestors: []string{
			"'none'",
		},
		FrameSrc: []string{
			"'none'",
		},
		ImgSrc: []string{
			"'self'",
			"data:",
			"https:",
		},
		ManifestSrc: []string{
			"'self'",
		},
		MediaSrc: []string{
			"'none'",
		},
		ObjectSrc: []string{
			"'none'",
		},
		// TODO: Re-enable. Issues on Chrome
		RequireTrustedTypesFor: []string{
			"'script'",
		},
		// 'unsafe-eval'
		// 'unsafe-inline' for backwards compatibility
		// 'self' OR 'strict-dynamic'
		// strict-dynamic does not allow host allowlisting
		//
		// script-src-elem => script-src
		ScriptSrc: []string{
			"'self'",
			"'nonce-RANDOM'",
			// NOTE: ONLY used for backwards-compatibility
			//       Browsers supporting nonce will ignore
			"'unsafe-inline'",
			// TODO: UNSAFE
			"'unsafe-eval'",
		},
		StyleSrc: []string{
			"'self'",
			"'nonce-RANDOM'",
			"fonts.googleapis.com",
		},
	}
}

func (csp *DataConfigCsp) Compile() string {
	return strings.Join([]string{
		fmt.Sprintf("default-src %s;", csp.DefaultSrc),
		fmt.Sprintf("base-uri %s;", csp.BaseUri),
		fmt.Sprintf("connect-src %s;", strings.Join(csp.ConnectSrc, " ")),
		fmt.Sprintf("font-src %s;", strings.Join(csp.FontSrc, " ")),
		fmt.Sprintf("form-action %s;", strings.Join(csp.FormAction, " ")),
		fmt.Sprintf("frame-ancestors %s;", strings.Join(csp.FrameAncestors, " ")),
		fmt.Sprintf("frame-src %s;", strings.Join(csp.FrameSrc, " ")),
		fmt.Sprintf("img-src %s;", strings.Join(csp.ImgSrc, " ")),
		fmt.Sprintf("manifest-src %s;", strings.Join(csp.ManifestSrc, " ")),
		fmt.Sprintf("media-src %s;", strings.Join(csp.MediaSrc, " ")),
		fmt.Sprintf("object-src %s;", strings.Join(csp.ObjectSrc, " ")),
		fmt.Sprintf("require-trusted-types-for %s;", strings.Join(csp.RequireTrustedTypesFor, " ")),
		fmt.Sprintf("script-src %s;", strings.Join(csp.ScriptSrc, " ")),
		fmt.Sprintf("style-src %s;", strings.Join(csp.StyleSrc, " ")),
	}, " ")
}

// func DefaultsEnv() map[string]string {
// 	return map[string]string{
// 		"LISTEN_ADDR":         "127.0.0.1",
// 		"LISTEN_PORT":         "8080",
// 	}
// }

func ConfigDefault() DataConfig {
	return DataConfig{
		Version:          1,
		ListenAddr:       "0.0.0.0",
		ListenPort:       "8080",
		NoncePlaceholder: "NONCEHERE",
		Banner: []string{
			`<!-- EZhttp ${BuildVersion} -->`,
		},
		Csp: DefaultConfigCsp(),
	}
}

func ConfigReadFromFile(filename string) DataConfig {

	// Read File
	filebytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("[CONFIG] Error when opening config file: ", err)
		return DataConfig{}
	}

	// Unroll
	var payload DataConfig
	err = json.Unmarshal(filebytes, &payload)
	if err != nil {
		log.Fatal("[CONFIG] Error in JSON File: ", err)
		return DataConfig{}
	}

	return payload
}

func ConfigLoad() DataConfig {
	const configfile string = "config.json"

	// Default Config
	c := ConfigDefault()

	// Check for Config File
	_, err := os.Stat(configfile)
	if os.IsNotExist(err) {
		log.Println("[CONFIG] File Not Found:", configfile)
	} else {
		log.Println("[CONFIG] File Found:", configfile)
		configFile := ConfigReadFromFile(configfile)
		errMerge := mergo.Merge(&c, configFile, mergo.WithOverride)
		if errMerge != nil {
			log.Println("[CONFIG] Merge Error", errMerge)
		}
	}

	// Environment Overrides for LISTEN/PORT
	envListen := os.Getenv("LISTEN")
	if envListen != "" {
		log.Println("[CONFIG] LISTEN OVERRIDE:", envListen)
		c.ListenAddr = envListen
	}
	envPort := os.Getenv("PORT")
	if envPort != "" {
		log.Println("[CONFIG] PORT OVERRIDE:", envPort)
		c.ListenPort = envPort
	}

	// Debug Log
	//log.Printf("%+v\n", c)

	// Return
	return c
}

// =========================
// =========================

func loadIndexCache() ([]string, error) {
	fileBytes, _ := os.ReadFile("./public/index.html")
	fileString := string(fileBytes)
	fileSplit := strings.Split(fileString, "NONCEHERE")
	if len(fileSplit) == 1 {
		log.Println("[INFO] No nonce field found. Use NONCEHERE in your file to use it")
	} else if len(fileSplit) == 2 {
		// All Good
		log.Println("[INFO] Found one nonce field")
	} else {
		// You probably do not need more than one nonce field
		// If you need it, you can remove this line and/or open a PR
		log.Println("[WARN] MORE THAN ONE NONCE FIELD FOUND")
	}

	//bufio.NewScanner()
	return fileSplit, nil
}

func generateIndexWithNonce(nonce string) string {
	totalSlots := len(cachedIndexString)
	finalReturn := make([]string, (totalSlots*2)-1)
	for i := 0; i < totalSlots; i++ {
		finalReturn[(i * 2)] = cachedIndexString[i]
		if (i + 1) != len(cachedIndexString) {
			finalReturn[(i*2)+1] = nonce
		}
	}
	finalString := strings.Join(finalReturn, "")
	minified, _ := minifier.String("text/html", finalString)
	minifiedWithBanner := strings.Replace(minified, "</head>", strings.Join(config.Banner, "\n")+"</head>", 1)
	minifiedAngularNonce := strings.Replace(minifiedWithBanner, "ngcspnonce", "ngCspNonce", 1)
	return minifiedAngularNonce
	//buf := &bytes.Buffer{}
	// import "encoding/gob"
	//gob.NewEncoder(buf).Encode(strings.Join(finalReturn, ""))
	//bs := buf.Bytes()
	//return bs
}

// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func HasDotPrefix(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("error:", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandStringCharacters(count int) string {
	setlength := len(charset)
	b := make([]byte, count)
	// Load random bytes (1 * b length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%setlength]
	}
	return string(b)
}

func checkFilePath(path string) (bool, string) {
	// FileStat
	exists := false
	_, errStatPath := os.Stat("./public" + path)
	if errStatPath == nil {
		exists = true
	} else {
		// Add HTML extension suffix
		_, errStatHtml := os.Stat("./public" + path + ".html")
		if errStatHtml == nil {
			exists = true
			path = path + ".html"
		}
	}
	return exists, path
}

func mwNonce(minhttpfs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract Path
		path := r.URL.Path
		//log.Println("PATH:", path)

		// Check for dotfiles
		// Double-dot (..) is handled before we even get control
		if HasDotPrefix(path) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "Forbidden")
			return
		}

		// Healthcheck
		if path == "/health" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK")
			return
		}

		// Redirect Root Index
		if path == "/index.html" || path == "/index.htm" || path == "/index" {
			// Redirect, Permanent
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}

		//lastChar, _ := utf8.DecodeLastRuneInString(path[:len(path)])
		lastChar := path[:]
		if path != "/" && lastChar == "/" {
			//log.Println("LAST CHAR IS / => /index.html")
			path = "/index.html"
		}

		// Global Headers
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Check for File
		pathexists, pathchecked := checkFilePath(path)
		//log.Println("CHECK PATH:", pathexists, pathchecked)

		// Doesn't Happen. Weird. We re-assign "/" to "/index.html" above to address.
		// The "/" will fall through to the static file server and be redirected.
		if path == "/" || !pathexists {
			//log.Println("SERVING INDEX:", path)

			// Check for possible 404
			// Allow setting known paths

			// Generate Nonce
			nonce := RandStringCharacters(32)

			// Write Response
			//_, _ = w.Write([]byte("BYTE"))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Security-Policy", strings.ReplaceAll(compiledCsp, "RANDOM", nonce))
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, generateIndexWithNonce(nonce))
			return
		} else {
			//log.Println("ACTUALLY EXISTS:", pathchecked)
			r.URL.Path = pathchecked

			cType, _, cacheable := getContentTypeByPath(pathchecked)
			w.Header().Set("Content-Type", cType)
			if cacheable {
				//log.Println("CACHE: YES")
				w.Header().Set("Cache-Control", "max-age=31536000, immutable")
			} else {
				//log.Println("CACHE: NO")
				w.Header().Set("Cache-Control", "no-cache")
			}
			// if minify {
			// 	minifier.Middleware(minhttpfs)
			// } else {
			//log.Println("ServeHTTP")
			minhttpfs.ServeHTTP(w, r)
			//}
		}
	}
}

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
// Note: Golang no longer requires an explicit "break" statement and is discouraged
func getContentTypeByPath(path string) (string, bool, bool) {
	splitted := strings.Split(path, ".")
	fileExt := splitted[len(splitted)-1]
	minify := false
	cacheable := false
	contentType := "application/octet-stream"
	switch fileExt {
	case "html":
		contentType = "text/html; charset=utf-8"
		minify = true
	case "css":
		contentType = "text/css; charset=utf-8"
		minify = false
		cacheable = true
	case "js":
		contentType = "text/javascript; charset=utf-8"
		minify = false
		cacheable = true
	case "json":
	case "map":
		contentType = "application/json; charset=utf-8"
		// TODO: Enable
		minify = false
	case "webmanifest":
		contentType = "application/manifest+json; charset=utf-8"
		// TODO: Enable
		minify = false
		cacheable = true
	case "ico":
		//contentType = "image/vnd.microsoft.icon"
		contentType = "image/x-icon"
		cacheable = true
	case "txt":
		contentType = "text/plain; charset=utf-8"
	case "jpeg":
	case "jpg":
		contentType = "image/jpeg"
		cacheable = true
	case "svg":
		contentType = "image/svg+xml"
		cacheable = true
	case "png":
		contentType = "image/png"
		cacheable = true
	case "weba":
		contentType = "audio/webm"
		cacheable = true
	case "webm":
		contentType = "video/webm"
		cacheable = true
	case "webp":
		contentType = "image/webp"
		cacheable = true
	case "woff2":
		contentType = "font/woff2; charset=utf-8"
	}
	return contentType, minify, cacheable
}

func PrintVersion() {
	fmt.Printf("v%s\nDate: %s\nGo Version: %s\nGit Hash: %s\n", BuildVersion, BuildDate, BuildGoVersion, BuildGitHash)
}

func PrintVersionShort() {
	fmt.Printf("### EZhttp v%s - Date: %s - Go Version: %s\n", BuildVersion, BuildDate, BuildGoVersion)
}

func main() {
	// Flag Check
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()
	if *showVersion {
		PrintVersion()
		return
	}

	PrintVersionShort()

	// Load Config
	config = ConfigLoad()

	// Cache Generated Index and CSP
	cachedIndexString, _ = loadIndexCache()
	compiledCsp = config.Csp.Compile()

	// Set Up Minification
	minifier = minify.New()
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

	// Static File Server
	httpfs := http.FileServer(http.Dir("./public"))
	// //http.Handle("/static/", http.StripPrefix("/public/", httpfs))
	// // strings.TrimRight("/statics/", "/")
	// http.Handle("/", mwNonce(httpfs))

	log.Printf("[SERVER] Listening on %s:%s...\n", config.ListenAddr, config.ListenPort)
	httpServer := &http.Server{
		Addr:              config.ListenAddr + ":" + config.ListenPort,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		//TLSConfig:         tlsConfig,
		Handler: mwNonce(httpfs),
	}
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
