package ratelimit

import (
	"net"
	"sync"
	"time"

	"github.com/ezhttp/ezhttp/internal/logger"
	"golang.org/x/time/rate"
)

// Manages per-IP rate limiting and blocking
type Limiter struct {
	mu               sync.Mutex
	limiters         map[string]*rate.Limiter
	limitersLastUsed map[string]time.Time
	blocked          map[string]time.Time // IP -> unblock time
	authFailures     map[string]int       // IP -> failure count
	requestsPerMin   int
	burstSize        int
	maxEntries       int // Maximum number of rate limiters to keep
	maxAuthAttempts  int
	blockDuration    time.Duration
}

// Creates a new rate limiter
func NewLimiter(requestsPerMin, burstSize int) *Limiter {
	l := &Limiter{
		limiters:         make(map[string]*rate.Limiter),
		limitersLastUsed: make(map[string]time.Time),
		blocked:          make(map[string]time.Time),
		authFailures:     make(map[string]int),
		requestsPerMin:   requestsPerMin,
		burstSize:        burstSize,
		maxEntries:       10000, // Default max entries
		maxAuthAttempts:  5,
		blockDuration:    15 * time.Minute,
	}

	return l
}

// Returns the rate limiter for the given IP
func (l *Limiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		// Check if we need to clean up before adding a new entry
		if len(l.limiters) >= l.maxEntries {
			l.cleanupOldest()
		}

		// Create new limiter: requests per minute converted to per second
		limiter = rate.NewLimiter(rate.Limit(float64(l.requestsPerMin)/60.0), l.burstSize)
		l.limiters[ip] = limiter
	}

	// Update last used time
	l.limitersLastUsed[ip] = time.Now()

	return limiter
}

// Checks if a request from the given IP is allowed
func (l *Limiter) Allow(ip string) bool {
	limiter := l.GetLimiter(ip)
	return limiter.Allow()
}

// Removes the oldest entries to make room for new ones
// Must be called with mutex held
func (l *Limiter) cleanupOldest() {
	// Find the 20% oldest entries to remove
	toRemove := l.maxEntries / 5
	if toRemove < 1 {
		toRemove = 1
	}

	// Create a slice of IPs sorted by last used time
	type ipTime struct {
		ip       string
		lastUsed time.Time
	}

	entries := make([]ipTime, 0, len(l.limitersLastUsed))
	for ip, lastUsed := range l.limitersLastUsed {
		entries = append(entries, ipTime{ip, lastUsed})
	}

	// Sort by last used time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].lastUsed.After(entries[j].lastUsed) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove the oldest entries
	removed := 0
	for _, entry := range entries {
		if removed >= toRemove {
			break
		}
		delete(l.limiters, entry.ip)
		delete(l.limitersLastUsed, entry.ip)
		removed++
	}

	if removed > 0 {
		logger.Debug("Cleaned up rate limiters due to size limit", "removed", removed, "remaining", len(l.limiters))
	}
}

// Extracts the client IP from the request
func ExtractIP(remoteAddr string) string {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If splitting fails, return the whole string
		return remoteAddr
	}
	return ip
}
