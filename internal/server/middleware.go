package server

import (
	"time"

	"github.com/ezhttp/ezhttp/internal/logger"
)

// Parses a duration string for cleanup interval
func ParseCleanupInterval(interval string) time.Duration {
	if interval == "" {
		return 30 * time.Minute // Default to 30 minutes
	}

	d, err := time.ParseDuration(interval)
	if err != nil {
		logger.Warn("Invalid cleanup interval, using default", "interval", interval, "error", err)
		return 30 * time.Minute
	}

	return d
}
