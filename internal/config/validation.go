package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Validates all configuration values
func ValidateConfig(c *DataConfig) error {
	// Validate listen address
	if err := validateListenAddr(c.ListenAddr); err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}

	// Validate port
	if err := validatePort(c.ListenPort); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// Validate nonce placeholder
	if err := validateNoncePlaceholder(c.NoncePlaceholder); err != nil {
		return fmt.Errorf("invalid nonce placeholder: %w", err)
	}

	// Validate CSP
	if err := validateCSP(&c.Csp); err != nil {
		return fmt.Errorf("invalid CSP configuration: %w", err)
	}

	return nil
}

// Validates the listen address format
func validateListenAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("listen address cannot be empty")
	}

	// Check for dangerous wildcards
	if addr == "*" {
		return fmt.Errorf("wildcard '*' not allowed, use '0.0.0.0' explicitly")
	}

	// Allow common valid addresses
	validAddresses := map[string]bool{
		"localhost": true,
		"127.0.0.1": true,
		"0.0.0.0":   true,
	}

	if validAddresses[addr] {
		return nil
	}

	// Validate as IP address
	if net.ParseIP(addr) == nil {
		// Try to resolve as hostname
		if _, err := net.LookupHost(addr); err != nil {
			return fmt.Errorf("invalid IP address or hostname: %s", addr)
		}
	}

	return nil
}

// Validates the port number
func validatePort(portStr string) error {
	if portStr == "" {
		return fmt.Errorf("port cannot be empty")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port must be a number: %s", portStr)
	}

	// Check port range
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	// Warn about privileged ports (but don't block)
	if port < 1024 {
		// This is just informational, not an error
		// User might be running with appropriate permissions
	}

	return nil
}

// Validates the nonce placeholder
func validateNoncePlaceholder(placeholder string) error {
	if placeholder == "" {
		return fmt.Errorf("nonce placeholder cannot be empty")
	}

	// Should not contain special characters that might break HTML/JS
	if strings.ContainsAny(placeholder, "<>\"'&") {
		return fmt.Errorf("nonce placeholder contains invalid characters")
	}

	// Should be reasonably short
	if len(placeholder) > 50 {
		return fmt.Errorf("nonce placeholder too long (max 50 chars)")
	}

	return nil
}

// Validates CSP configuration
func validateCSP(csp *DataConfigCsp) error {
	// Validate required CSP directives are not empty
	if csp.DefaultSrc == "" {
		return fmt.Errorf("default-src cannot be empty")
	}

	// Check for dangerous CSP values
	dangerousValues := []string{
		"*",      // Allow all origins is dangerous
		"data:*", // Unrestricted data URIs
		"blob:*", // Unrestricted blob URIs
	}

	// Helper function to check directive values
	checkDirective := func(name string, values []string) error {
		for _, value := range values {
			// Check for dangerous wildcards
			for _, dangerous := range dangerousValues {
				if value == dangerous {
					return fmt.Errorf("%s contains dangerous value: %s", name, value)
				}
			}

			// Basic validation of value format
			if strings.Contains(value, " ") && !strings.HasPrefix(value, "'") {
				return fmt.Errorf("%s contains unquoted value with space: %s", name, value)
			}
		}
		return nil
	}

	// Validate each directive
	if err := checkDirective("connect-src", csp.ConnectSrc); err != nil {
		return err
	}
	if err := checkDirective("font-src", csp.FontSrc); err != nil {
		return err
	}
	if err := checkDirective("form-action", csp.FormAction); err != nil {
		return err
	}
	if err := checkDirective("frame-ancestors", csp.FrameAncestors); err != nil {
		return err
	}
	if err := checkDirective("frame-src", csp.FrameSrc); err != nil {
		return err
	}
	if err := checkDirective("img-src", csp.ImgSrc); err != nil {
		return err
	}
	if err := checkDirective("manifest-src", csp.ManifestSrc); err != nil {
		return err
	}
	if err := checkDirective("media-src", csp.MediaSrc); err != nil {
		return err
	}
	if err := checkDirective("object-src", csp.ObjectSrc); err != nil {
		return err
	}
	if err := checkDirective("script-src", csp.ScriptSrc); err != nil {
		return err
	}
	if err := checkDirective("style-src", csp.StyleSrc); err != nil {
		return err
	}

	// Warn about unsafe CSP values (but don't block)
	for _, scriptSrc := range csp.ScriptSrc {
		if scriptSrc == "'unsafe-inline'" || scriptSrc == "'unsafe-eval'" {
			// This is logged during ConfigLoad, not an error
			// as it might be intentionally needed
		}
	}

	return nil
}
