package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// HasDotPrefix reports whether name contains a path element starting with a period.
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

func CheckFilePath(path string) (bool, string) {
	cleanPath := filepath.Clean(path)

	if strings.Contains(cleanPath, "..") {
		return false, ""
	}

	basePath, err := filepath.Abs("./public")
	if err != nil {
		return false, ""
	}

	fullPath := filepath.Join(basePath, cleanPath)

	if !strings.HasPrefix(fullPath, basePath) {
		return false, ""
	}

	fileInfo, err := os.Stat(fullPath)
	if err == nil {
		if fileInfo.Mode().IsRegular() {
			return true, cleanPath
		}
		return false, ""
	}

	htmlPath := fullPath + ".html"
	fileInfo, err = os.Stat(htmlPath)
	if err == nil && fileInfo.Mode().IsRegular() {
		return true, cleanPath + ".html"
	}

	return false, ""
}
