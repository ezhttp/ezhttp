package server

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ezhttp/ezhttp/internal/logger"
)

// FileExistenceCache caches file existence checks to avoid redundant stat calls
type FileExistenceCache struct {
	mu       sync.RWMutex
	cache    map[string]*cacheEntry
	basePath string
	ttl      time.Duration
}

type cacheEntry struct {
	exists       bool
	resolvedPath string
	timestamp    time.Time
}

// NewFileExistenceCache creates a new file existence cache
func NewFileExistenceCache(basePath string, ttl time.Duration) (*FileExistenceCache, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	cache := &FileExistenceCache{
		cache:    make(map[string]*cacheEntry),
		basePath: absPath,
		ttl:      ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache, nil
}

// CheckPath checks if a path exists, using cache when possible
func (c *FileExistenceCache) CheckPath(path string) (exists bool, resolvedPath string) {
	// Security check - check BEFORE cleaning
	if strings.Contains(path, "..") {
		return false, ""
	}

	cleanPath := filepath.Clean(path)

	// Check cache first
	c.mu.RLock()
	entry, found := c.cache[cleanPath]
	if found && time.Since(entry.timestamp) < c.ttl {
		c.mu.RUnlock()
		return entry.exists, entry.resolvedPath
	}
	c.mu.RUnlock()

	// Cache miss or expired, check file system
	exists, resolvedPath = c.checkFileSystem(cleanPath)

	// Update cache
	c.mu.Lock()
	c.cache[cleanPath] = &cacheEntry{
		exists:       exists,
		resolvedPath: resolvedPath,
		timestamp:    time.Now(),
	}
	c.mu.Unlock()

	return exists, resolvedPath
}

// checkFileSystem performs the actual file system check
func (c *FileExistenceCache) checkFileSystem(cleanPath string) (bool, string) {
	fullPath := filepath.Join(c.basePath, cleanPath)

	// Security check
	if !strings.HasPrefix(fullPath, c.basePath) {
		return false, ""
	}

	// Check exact path
	fileInfo, err := os.Stat(fullPath)
	if err == nil {
		if fileInfo.Mode().IsRegular() {
			return true, cleanPath
		}
		return false, ""
	}

	// Check with .html extension
	htmlPath := fullPath + ".html"
	fileInfo, err = os.Stat(htmlPath)
	if err == nil && fileInfo.Mode().IsRegular() {
		return true, cleanPath + ".html"
	}

	return false, ""
}

// cleanupLoop periodically removes expired entries
func (c *FileExistenceCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes expired cache entries
func (c *FileExistenceCache) cleanupExpired() {
	now := time.Now()
	expired := make([]string, 0)

	c.mu.RLock()
	for path, entry := range c.cache {
		if now.Sub(entry.timestamp) > c.ttl {
			expired = append(expired, path)
		}
	}
	c.mu.RUnlock()

	if len(expired) > 0 {
		c.mu.Lock()
		for _, path := range expired {
			delete(c.cache, path)
		}
		c.mu.Unlock()
		logger.Debug("Cleaned up expired cache entries", "count", len(expired))
	}
}

// PrewarmCommonPaths pre-populates the cache with common static file paths
func (c *FileExistenceCache) PrewarmCommonPaths() {
	commonPaths := []string{
		"/favicon.ico",
		"/robots.txt",
		"/sitemap.xml",
		"/manifest.json",
		"/service-worker.js",
		"/sw.js",
		"/apple-touch-icon.png",
		"/apple-touch-icon-precomposed.png",
	}

	// Also check common asset directories
	assetDirs := []string{
		"/css",
		"/js",
		"/assets",
		"/images",
		"/img",
		"/fonts",
		"/static",
	}

	// Check common paths
	for _, path := range commonPaths {
		c.CheckPath(path)
	}

	// For directories, check if index.html exists
	for _, dir := range assetDirs {
		c.CheckPath(dir + "/index.html")
	}

	logger.Debug("Pre-warmed file cache", "paths", len(commonPaths)+len(assetDirs))
}
