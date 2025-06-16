package server

import (
	"crypto/tls"
)

// Creates a secure TLS configuration
func CreateTLSConfig() *tls.Config {
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
