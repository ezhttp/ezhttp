package utils

import "strings"

// GetContentTypeByPath returns content type, minify flag, and cacheable flag based on file extension
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
// Note: Golang no longer requires an explicit "break" statement and is discouraged
func GetContentTypeByPath(path string) (string, bool, bool) {
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
