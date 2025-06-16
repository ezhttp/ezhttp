package tls

import (
	"crypto/tls"

	"github.com/ezhttp/ezhttp/internal/logger"
)

// TLS security configuration level
type SecurityLevel int

const (
	// SecurityLevelStrong uses only the most secure cipher suites
	SecurityLevelStrong SecurityLevel = iota
	// SecurityLevelRelaxed uses default cipher suites for compatibility
	SecurityLevelRelaxed
	// SecurityLevelInsecure skips certificate verification (dangerous!)
	SecurityLevelInsecure
)

// Creates a secure TLS configuration for servers
func CreateServerTLSConfig() *tls.Config {
	return &tls.Config{
		// Minimum TLS version
		MinVersion: tls.VersionTLS12,

		// Strong cipher suites only
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites (automatically preferred when available)
			// TLS 1.2 cipher suites (ECDHE + AEAD only)
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},

		// Prefer server cipher suites for TLS 1.2
		PreferServerCipherSuites: true,

		// Use modern curves only
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}
}

// Creates a TLS configuration for clients (proxy to backend)
func CreateClientTLSConfig(level SecurityLevel) *tls.Config {
	switch level {
	case SecurityLevelInsecure:
		// Skip certificate verification (dangerous, but needed for some legacy systems)
		logger.Warn("TLS configured to skip certificate verification")
		return &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}

	case SecurityLevelRelaxed:
		// Use default cipher suites for compatibility
		logger.Info("Using relaxed TLS settings")
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

	case SecurityLevelStrong:
		fallthrough
	default:
		// Strong TLS configuration
		logger.Info("Using strong TLS settings")
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			},
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
			},
		}
	}
}
