package security

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// AllowedExtensions defines which file extensions can be served
var AllowedExtensions = map[string]bool{
	".html":        true,
	".htm":         true,
	".css":         true,
	".js":          true,
	".json":        true,
	".webmanifest": true,
	".txt":         true,
	".ico":         true,
	".jpg":         true,
	".jpeg":        true,
	".png":         true,
	".gif":         true,
	".svg":         true,
	".webp":        true,
	".woff":        true,
	".woff2":       true,
	".ttf":         true,
	".eot":         true,
	".otf":         true,
	".mp4":         true,
	".webm":        true,
	".weba":        true,
	".ogg":         true,
	".mp3":         true,
	".wav":         true,
	".pdf":         true,
	".map":         true,
}

// SecureFileSystem implements http.FileSystem with additional security checks
type SecureFileSystem struct {
	Fs http.FileSystem
}

// Open implements http.FileSystem with security validations
func (sfs *SecureFileSystem) Open(name string) (http.File, error) {
	// Clean the path
	cleanPath := filepath.Clean("/" + name)

	// Prevent directory traversal
	if strings.Contains(cleanPath, "..") {
		return nil, os.ErrNotExist
	}

	// Open the file
	file, err := sfs.Fs.Open(cleanPath)
	if err != nil {
		return nil, err
	}

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	// Prevent directory listings
	if stat.IsDir() {
		indexPath := filepath.Join(cleanPath, "index.html")
		if indexFile, err := sfs.Fs.Open(indexPath); err == nil {
			indexFile.Close()
			file.Close()
			// Redirect to index.html
			return sfs.Fs.Open(indexPath)
		}
		// No index.html, deny directory listing
		file.Close()
		return nil, os.ErrNotExist
	}

	// Check if it's a symlink (additional security)
	if stat.Mode()&os.ModeSymlink != 0 {
		file.Close()
		return nil, os.ErrPermission
	}

	// Validate file extension
	ext := filepath.Ext(cleanPath)
	if ext != "" && !AllowedExtensions[strings.ToLower(ext)] {
		file.Close()
		return nil, os.ErrPermission
	}

	return file, nil
}
