package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dario.cat/mergo"
	"github.com/ezhttp/ezhttp/internal/logger"
)

type DataConfig struct {
	Version          int                 `json:"version"`
	ListenAddr       string              `json:"listen_addr"`
	ListenPort       string              `json:"listen_port"`
	NoncePlaceholder string              `json:"nonce_placeholder"`
	Banner           []string            `json:"banner"`
	Csp              DataConfigCsp       `json:"csp"`
	RateLimit        DataConfigRateLimit `json:"rate_limit"`
	TLS              DataConfigTLS       `json:"tls"`
	Proxy            DataConfigProxy     `json:"proxy"`
}

type DataConfigCsp struct {
	DefaultSrc             string   `json:"default-src"`
	BaseUri                string   `json:"base-uri"`
	ConnectSrc             []string `json:"connect-src"`
	FontSrc                []string `json:"font-src"`
	FormAction             []string `json:"form-action"`
	FrameAncestors         []string `json:"frame-ancestors"`
	FrameSrc               []string `json:"frame-src"`
	ImgSrc                 []string `json:"img-src"`
	ManifestSrc            []string `json:"manifest-src"`
	MediaSrc               []string `json:"media-src"`
	ObjectSrc              []string `json:"object-src"`
	RequireTrustedTypesFor []string `json:"require-trusted-types-for"`
	ScriptSrc              []string `json:"script-src"`
	StyleSrc               []string `json:"style-src"`
}

type DataConfigRateLimit struct {
	Enabled           bool   `json:"enabled"`
	RequestsPerMinute int    `json:"requests_per_minute"`
	BurstSize         int    `json:"burst_size"`
	CleanupInterval   string `json:"cleanup_interval"`
}

type DataConfigTLS struct {
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

type DataConfigProxy struct {
	OriginBaseURL          string `json:"origin_base_url"`
	AuthToken              string `json:"auth_token"`
	AllowedHost            string `json:"allowed_host"`
	AllowInsecureOriginTLS bool   `json:"allow_insecure_origin_tls"`
	RelaxedOriginTLS       bool   `json:"relaxed_origin_tls"`
	MaxIdleConns           int    `json:"max_idle_conns"`
	IdleConnTimeout        string `json:"idle_conn_timeout"`
	MaxRequestSize         int64  `json:"max_request_size"`
	MaxAuthAttempts        int    `json:"max_auth_attempts"`
	BlockDuration          string `json:"block_duration"`
	DebugMode              bool   `json:"debug_mode"`
}

func DefaultConfigCsp() DataConfigCsp {
	// sandbox
	// TODO: Setup
	//"report-uri https://csp.example.com;",
	// child-src
	//"report-to https://csp.example.com;",
	//"prefetch-src 'none'",
	// navigate-to
	// upgrade-insecure-requests
	// block-all-mixed-content
	// Do not use. Removed
	//"plugin-types 'none';",
	return DataConfigCsp{
		// TODO: Move to 'none'. Angular breaks for some reason
		DefaultSrc: "'self'",
		BaseUri:    "'self'",
		ConnectSrc: []string{
			"'self'",
			"https://fonts.gstatic.com",
		},
		FontSrc: []string{
			"'self'",
			"fonts.gstatic.com",
		},
		FormAction: []string{
			"'self'",
		},
		FrameAncestors: []string{
			"'none'",
		},
		FrameSrc: []string{
			"'none'",
		},
		ImgSrc: []string{
			"'self'",
			"data:",
			"https:",
		},
		ManifestSrc: []string{
			"'self'",
		},
		MediaSrc: []string{
			"'none'",
		},
		ObjectSrc: []string{
			"'none'",
		},
		// TODO: Re-enable. Issues on Chrome
		RequireTrustedTypesFor: []string{
			"'script'",
		},
		// 'unsafe-eval'
		// 'unsafe-inline' for backwards compatibility
		// 'self' OR 'strict-dynamic'
		// strict-dynamic does not allow host allowlisting
		//
		// script-src-elem => script-src
		ScriptSrc: []string{
			"'self'",
			"'nonce-RANDOM'",
			// NOTE: ONLY used for backwards-compatibility
			//       Browsers supporting nonce will ignore
			"'unsafe-inline'",
			// TODO: UNSAFE
			"'unsafe-eval'",
		},
		StyleSrc: []string{
			"'self'",
			"'nonce-RANDOM'",
			"fonts.googleapis.com",
		},
	}
}

func (csp *DataConfigCsp) Compile() string {
	return strings.Join([]string{
		fmt.Sprintf("default-src %s;", csp.DefaultSrc),
		fmt.Sprintf("base-uri %s;", csp.BaseUri),
		fmt.Sprintf("connect-src %s;", strings.Join(csp.ConnectSrc, " ")),
		fmt.Sprintf("font-src %s;", strings.Join(csp.FontSrc, " ")),
		fmt.Sprintf("form-action %s;", strings.Join(csp.FormAction, " ")),
		fmt.Sprintf("frame-ancestors %s;", strings.Join(csp.FrameAncestors, " ")),
		fmt.Sprintf("frame-src %s;", strings.Join(csp.FrameSrc, " ")),
		fmt.Sprintf("img-src %s;", strings.Join(csp.ImgSrc, " ")),
		fmt.Sprintf("manifest-src %s;", strings.Join(csp.ManifestSrc, " ")),
		fmt.Sprintf("media-src %s;", strings.Join(csp.MediaSrc, " ")),
		fmt.Sprintf("object-src %s;", strings.Join(csp.ObjectSrc, " ")),
		fmt.Sprintf("require-trusted-types-for %s;", strings.Join(csp.RequireTrustedTypesFor, " ")),
		fmt.Sprintf("script-src %s;", strings.Join(csp.ScriptSrc, " ")),
		fmt.Sprintf("style-src %s;", strings.Join(csp.StyleSrc, " ")),
	}, " ")
}

func ConfigDefault() DataConfig {
	return DataConfig{
		Version:          1,
		ListenAddr:       "127.0.0.1",
		ListenPort:       "8080",
		NoncePlaceholder: "NONCEHERE",
		Banner: []string{
			`<!-- EZhttp ${BuildVersion} -->`,
		},
		Csp: DefaultConfigCsp(),
		RateLimit: DataConfigRateLimit{
			Enabled:           true,
			RequestsPerMinute: 60,
			BurstSize:         10,
			CleanupInterval:   "30m",
		},
		TLS: DataConfigTLS{
			CertFile: "",
			KeyFile:  "",
		},
		Proxy: DataConfigProxy{
			OriginBaseURL:          "",
			AuthToken:              "",
			AllowedHost:            "",
			AllowInsecureOriginTLS: false,
			RelaxedOriginTLS:       false,
			MaxIdleConns:           100,
			IdleConnTimeout:        "90s",
			MaxRequestSize:         52428800, // 50MB
			MaxAuthAttempts:        5,
			BlockDuration:          "15m",
			DebugMode:              false,
		},
	}
}

func ConfigReadFromFile(filename string) DataConfig {
	filebytes, err := os.ReadFile(filename)
	if err != nil {
		logger.Fatal("Error opening config file", "error", err)
		return DataConfig{}
	}

	var payload DataConfig
	err = json.Unmarshal(filebytes, &payload)
	if err != nil {
		logger.Fatal("Error parsing JSON config", "error", err)
		return DataConfig{}
	}

	return payload
}

func ConfigLoad() DataConfig {
	const configfile string = "config.json"

	c := ConfigDefault()

	_, err := os.Stat(configfile)
	if os.IsNotExist(err) {
		logger.Info("Config file not found, using defaults", "file", configfile)
	} else {
		logger.Info("Loading config file", "file", configfile)
		configFile := ConfigReadFromFile(configfile)
		errMerge := mergo.Merge(&c, configFile, mergo.WithOverride)
		if errMerge != nil {
			logger.Error("Failed to merge config", "error", errMerge)
		}
	}

	envListen := os.Getenv("LISTEN")
	if envListen != "" {
		logger.Info("Environment override for listen address", "address", envListen)
		c.ListenAddr = envListen
	}
	envPort := os.Getenv("PORT")
	if envPort != "" {
		logger.Info("Environment override for port", "port", envPort)
		c.ListenPort = envPort
	}

	// TLS environment overrides
	envTLSCert := os.Getenv("TLS_CERT")
	if envTLSCert != "" {
		logger.Info("Environment override for TLS certificate", "path", envTLSCert)
		c.TLS.CertFile = envTLSCert
	}
	envTLSKey := os.Getenv("TLS_KEY")
	if envTLSKey != "" {
		logger.Info("Environment override for TLS key", "path", envTLSKey)
		c.TLS.KeyFile = envTLSKey
	}

	// Proxy environment overrides
	envProxyTarget := os.Getenv("PROXY_TARGET")
	if envProxyTarget != "" {
		logger.Info("Environment override for proxy target", "url", envProxyTarget)
		c.Proxy.OriginBaseURL = envProxyTarget
	}
	envProxyAuth := os.Getenv("PROXY_AUTH_TOKEN")
	if envProxyAuth != "" {
		logger.Info("Environment override for proxy auth token")
		c.Proxy.AuthToken = envProxyAuth
	}
	envAllowedHost := os.Getenv("ALLOWED_HOST")
	if envAllowedHost != "" {
		logger.Info("Environment override for allowed host", "host", envAllowedHost)
		c.Proxy.AllowedHost = envAllowedHost
	}
	envDebugMode := os.Getenv("DEBUG_MODE")
	if envDebugMode == "true" {
		logger.Info("Debug mode enabled")
		c.Proxy.DebugMode = true
	}

	// Validate configuration
	if err := ValidateConfig(&c); err != nil {
		logger.Fatal("Config validation failed", "error", err)
	}

	// Log security warnings
	if c.ListenAddr == "0.0.0.0" {
		logger.Warn("Server will listen on all network interfaces")
	}

	// Check for unsafe CSP directives
	for _, scriptSrc := range c.Csp.ScriptSrc {
		if scriptSrc == "'unsafe-inline'" {
			logger.Warn("'unsafe-inline' in script-src reduces XSS protection")
		}
		if scriptSrc == "'unsafe-eval'" {
			logger.Warn("'unsafe-eval' in script-src allows dynamic code execution")
		}
	}

	return c
}
