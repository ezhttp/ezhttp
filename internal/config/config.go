package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"dario.cat/mergo"
)

type DataConfig struct {
	Version          int           `json:"version"`
	ListenAddr       string        `json:"listen_addr"`
	ListenPort       string        `json:"listen_port"`
	NoncePlaceholder string        `json:"nonce_placeholder"`
	Banner           []string      `json:"banner"`
	Csp              DataConfigCsp `json:"csp"`
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

// func DefaultsEnv() map[string]string {
// 	return map[string]string{
// 		"LISTEN_ADDR":         "127.0.0.1",
// 		"LISTEN_PORT":         "8080",
// 	}
// }

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
	}
}

func ConfigReadFromFile(filename string) DataConfig {
	filebytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("[CONFIG] Error when opening config file")
		return DataConfig{}
	}

	var payload DataConfig
	err = json.Unmarshal(filebytes, &payload)
	if err != nil {
		log.Fatal("[CONFIG] Error in JSON file")
		return DataConfig{}
	}

	return payload
}

func ConfigLoad() DataConfig {
	const configfile string = "config.json"

	c := ConfigDefault()

	_, err := os.Stat(configfile)
	if os.IsNotExist(err) {
		log.Println("[CONFIG] File Not Found:", configfile)
	} else {
		log.Println("[CONFIG] File Found:", configfile)
		configFile := ConfigReadFromFile(configfile)
		errMerge := mergo.Merge(&c, configFile, mergo.WithOverride)
		if errMerge != nil {
			log.Println("[CONFIG] Merge Error", errMerge)
		}
	}

	envListen := os.Getenv("LISTEN")
	if envListen != "" {
		log.Println("[CONFIG] LISTEN OVERRIDE:", envListen)
		c.ListenAddr = envListen
	}
	envPort := os.Getenv("PORT")
	if envPort != "" {
		log.Println("[CONFIG] PORT OVERRIDE:", envPort)
		c.ListenPort = envPort
	}

	// Validate configuration
	if err := ValidateConfig(&c); err != nil {
		log.Fatal("[CONFIG] Validation failed: ", err)
	}

	// Log security warnings
	if c.ListenAddr == "0.0.0.0" {
		log.Println("[CONFIG] WARNING: Server will listen on all network interfaces")
	}

	// Check for unsafe CSP directives
	for _, scriptSrc := range c.Csp.ScriptSrc {
		if scriptSrc == "'unsafe-inline'" {
			log.Println("[CONFIG] WARNING: 'unsafe-inline' in script-src reduces XSS protection")
		}
		if scriptSrc == "'unsafe-eval'" {
			log.Println("[CONFIG] WARNING: 'unsafe-eval' in script-src allows dynamic code execution")
		}
	}

	return c
}
