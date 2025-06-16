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
	cleanupInterval  time.Duration
	maxAuthAttempts  int
	blockDuration    time.Duration
}

// Creates a new rate limiter
func NewLimiter(requestsPerMin, burstSize int, cleanupInterval time.Duration) *Limiter {
	l := &Limiter{
		limiters:         make(map[string]*rate.Limiter),
		limitersLastUsed: make(map[string]time.Time),
		blocked:          make(map[string]time.Time),
		authFailures:     make(map[string]int),
		requestsPerMin:   requestsPerMin,
		burstSize:        burstSize,
		cleanupInterval:  cleanupInterval,
		maxAuthAttempts:  5,
		blockDuration:    15 * time.Minute,
	}

	// Start cleanup goroutine
	if cleanupInterval > 0 {
		go l.cleanupRoutine()
	}

	return l
}

// Returns the rate limiter for the given IP
func (l *Limiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
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

// Periodically removes unused limiters
func (l *Limiter) cleanupRoutine() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.cleanup()
	}
}

// Removes limiters that haven't been used recently
func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour) // Remove limiters unused for 1 hour

	for ip, lastUsed := range l.limitersLastUsed {
		if lastUsed.Before(cutoff) {
			delete(l.limiters, ip)
			delete(l.limitersLastUsed, ip)
			logger.Debug("Cleaned up rate limiter", "ip", ip)
		}
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
