package logger

import (
	"context"
	"log/slog"
	"os"
)

var DefaultLogger *slog.Logger

func init() {
	// Set log level based on environment variable
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Redact sensitive information
			switch a.Key {
			case "error":
				// Sanitize error messages
				if err, ok := a.Value.Any().(error); ok {
					a.Value = slog.StringValue(sanitizeError(err))
				}
			case "path", "file", "filename":
				// Show only filename, not full path
				if str, ok := a.Value.Any().(string); ok {
					a.Value = slog.StringValue(sanitizePath(str))
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	DefaultLogger = slog.New(handler)
	slog.SetDefault(DefaultLogger)
}

// Removes potentially sensitive information from errors
func sanitizeError(err error) string {
	// Return generic error messages
	errStr := err.Error()

	// Common error types that might leak information
	switch {
	case contains(errStr, "permission denied"):
		return "permission denied"
	case contains(errStr, "no such file"):
		return "file not found"
	case contains(errStr, "connection refused"):
		return "connection error"
	case contains(errStr, "timeout"):
		return "operation timeout"
	default:
		// For unknown errors, return a generic message
		return "internal error"
	}
}

// Returns only the filename without the full path
func sanitizePath(path string) string {
	// Find the last slash
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsAt(s, substr)
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Info logs at INFO level with structured fields
func Info(msg string, args ...any) {
	DefaultLogger.Info(msg, args...)
}

// Error logs at ERROR level with structured fields
func Error(msg string, args ...any) {
	DefaultLogger.Error(msg, args...)
}

// Warn logs at WARN level with structured fields
func Warn(msg string, args ...any) {
	DefaultLogger.Warn(msg, args...)
}

// Debug logs at DEBUG level with structured fields
func Debug(msg string, args ...any) {
	DefaultLogger.Debug(msg, args...)
}

// Fatal logs at ERROR level and then calls os.Exit(1)
func Fatal(msg string, args ...any) {
	DefaultLogger.Error(msg, args...)
	os.Exit(1)
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) *slog.Logger {
	return DefaultLogger.With("context", ctx)
}
