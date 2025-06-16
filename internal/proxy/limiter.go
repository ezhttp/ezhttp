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
	cleanupInterval  time.Duration
	maxAuthAttempts  int
	blockDuration    time.Duration
}

// NewProxyLimiter creates a new proxy rate limiter with IP blocking
func NewProxyLimiter(requestsPerMin, burstSize, maxAuthAttempts int, cleanupInterval, blockDuration time.Duration) *ProxyLimiter {
	l := &ProxyLimiter{
		limiters:         make(map[string]*rate.Limiter),
		limitersLastUsed: make(map[string]time.Time),
		blocked:          make(map[string]time.Time),
		authFailures:     make(map[string]int),
		requestsPerMin:   requestsPerMin,
		burstSize:        burstSize,
		cleanupInterval:  cleanupInterval,
		maxAuthAttempts:  maxAuthAttempts,
		blockDuration:    blockDuration,
	}

	// Start cleanup goroutine
	if cleanupInterval > 0 {
		go l.cleanupRoutine()
	}

	return l
}

// GetLimiter returns the rate limiter for the given IP
func (l *ProxyLimiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
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

// cleanupRoutine periodically removes expired entries
func (l *ProxyLimiter) cleanupRoutine() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.cleanup()
	}
}

// cleanup removes expired blocks and unused limiters
func (l *ProxyLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Clean up expired blocks
	for ip, unblockTime := range l.blocked {
		if now.After(unblockTime) {
			delete(l.blocked, ip)
			delete(l.authFailures, ip)
			logger.Debug("Cleaned up expired block", "ip", ip)
		}
	}

	// Clean up unused rate limiters
	cutoff := now.Add(-1 * time.Hour)
	for ip, lastUsed := range l.limitersLastUsed {
		if lastUsed.Before(cutoff) {
			delete(l.limiters, ip)
			delete(l.limitersLastUsed, ip)
			logger.Debug("Cleaned up unused rate limiter", "ip", ip)
		}
	}
}
