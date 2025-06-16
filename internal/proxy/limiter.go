package proxy

import (
	"sync"
	"time"

	"github.com/ezhttp/ezhttp/internal/logger"
	"golang.org/x/time/rate"
)

// ProxyLimiter extends the basic rate limiter with IP blocking capabilities
type ProxyLimiter struct {
	mu               sync.Mutex
	limiters         map[string]*rate.Limiter
	limitersLastUsed map[string]time.Time
	blocked          map[string]time.Time
	authFailures     map[string]int
	requestsPerMin   int
	burstSize        int
	maxEntries       int // Maximum number of rate limiters to keep
	maxAuthAttempts  int
	blockDuration    time.Duration
}

// NewProxyLimiter creates a new proxy rate limiter with IP blocking
func NewProxyLimiter(requestsPerMin, burstSize, maxAuthAttempts int, blockDuration time.Duration) *ProxyLimiter {
	l := &ProxyLimiter{
		limiters:         make(map[string]*rate.Limiter),
		limitersLastUsed: make(map[string]time.Time),
		blocked:          make(map[string]time.Time),
		authFailures:     make(map[string]int),
		requestsPerMin:   requestsPerMin,
		burstSize:        burstSize,
		maxEntries:       10000, // Default max entries
		maxAuthAttempts:  maxAuthAttempts,
		blockDuration:    blockDuration,
	}

	return l
}

// GetLimiter returns the rate limiter for the given IP
func (l *ProxyLimiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		// Check if we need to clean up before adding a new entry
		if len(l.limiters) >= l.maxEntries {
			l.cleanupOldest()
		}

		limiter = rate.NewLimiter(rate.Limit(float64(l.requestsPerMin)/60.0), l.burstSize)
		l.limiters[ip] = limiter
	}

	l.limitersLastUsed[ip] = time.Now()
	return limiter
}

// Allow checks if a request from the given IP is allowed
func (l *ProxyLimiter) Allow(ip string) bool {
	// Check if IP is blocked first
	if l.IsBlocked(ip) {
		return false
	}

	limiter := l.GetLimiter(ip)
	return limiter.Allow()
}

// IsBlocked checks if an IP is currently blocked
func (l *ProxyLimiter) IsBlocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	unblockTime, exists := l.blocked[ip]
	if !exists {
		return false
	}

	// Check if block has expired
	if time.Now().After(unblockTime) {
		delete(l.blocked, ip)
		delete(l.authFailures, ip)
		logger.Info("IP unblocked", "ip", ip)
		return false
	}

	return true
}

// RecordAuthFailure records a failed authentication attempt
func (l *ProxyLimiter) RecordAuthFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.authFailures[ip]++
	failures := l.authFailures[ip]

	logger.Warn("Authentication failure",
		"ip", ip,
		"failures", failures,
		"max_attempts", l.maxAuthAttempts)

	if failures >= l.maxAuthAttempts {
		l.blocked[ip] = time.Now().Add(l.blockDuration)
		logger.Warn("IP blocked due to auth failures",
			"ip", ip,
			"duration", l.blockDuration.String())
	}
}

// ResetAuthFailures resets the failure count for an IP (on successful auth)
func (l *ProxyLimiter) ResetAuthFailures(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.authFailures, ip)
}

// cleanupOldest removes the oldest entries to make room for new ones
// Must be called with mutex held
func (l *ProxyLimiter) cleanupOldest() {
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
