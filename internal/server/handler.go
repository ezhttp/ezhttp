package server

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/ezhttp/ezhttp/internal/utils"
	"github.com/tdewolff/minify/v2"
)

// MwNonce is the main HTTP handler middleware that adds nonces and serves files
func MwNonce(minhttpfs http.Handler, compiledCsp string, cachedIndexString []string, minifier *minify.M, banner []string, fileCache *FileExistenceCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract Path
		path := r.URL.Path
		//log.Println("PATH:", path)

		// Sanitize path
		cleanPath := filepath.Clean(path)

		// Check for dotfiles and path traversal attempts
		if utils.HasDotPrefix(cleanPath) || strings.Contains(cleanPath, "..") {
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

		// Check for File using cache
		pathexists, pathchecked := fileCache.CheckPath(path)
		//log.Println("CHECK PATH:", pathexists, pathchecked)

		// Doesn't Happen. Weird. We re-assign "/" to "/index.html" above to address.
		// The "/" will fall through to the static file server and be redirected.
		if path == "/" || !pathexists {
			//log.Println("SERVING INDEX:", path)

			// Check for possible 404
			// Allow setting known paths

			// Generate Nonce or Fail
			nonce := utils.RandStringCharacters(32)
			if nonce == "" {
				// Fail if the nonce generated incorrectly
				logger.Error("Failed to generate secure nonce", "type", "security")
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "Internal Server Error")
				return
			}

			// Write Response
			//_, _ = w.Write([]byte("BYTE"))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Security-Policy", strings.ReplaceAll(compiledCsp, "RANDOM", nonce))
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, GenerateIndexWithNonce(nonce, cachedIndexString, minifier, banner))
			return
		} else {
			//log.Println("ACTUALLY EXISTS:", pathchecked)
			r.URL.Path = pathchecked

			cType, _, cacheable := utils.GetContentTypeByPath(pathchecked)
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
