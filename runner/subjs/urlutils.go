package subjs

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidateAndNormalizeURL performs comprehensive URL validation and normalization
func ValidateAndNormalizeURL(inputURL string) (*url.URL, error) {
	if inputURL == "" {
		return nil, URLValidationError{URL: inputURL, Reason: "empty URL"}
	}

	// Parse the URL
	parsed, err := url.Parse(strings.TrimSpace(inputURL))
	if err != nil {
		return nil, URLValidationError{URL: inputURL, Reason: fmt.Sprintf("invalid URL format: %v", err)}
	}

	// Ensure scheme is present and valid
	if parsed.Scheme == "" {
		return nil, URLValidationError{URL: inputURL, Reason: "missing URL scheme"}
	}

	// Only allow http/https schemas
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, URLValidationError{URL: inputURL, Reason: fmt.Sprintf("unsupported scheme '%s', only http/https allowed", parsed.Scheme)}
	}

	// Ensure host is present
	if parsed.Host == "" {
		return nil, URLValidationError{URL: inputURL, Reason: "missing host"}
	}

	// Validate hostname
	if err := validateHostname(parsed.Host); err != nil {
		return nil, URLValidationError{URL: inputURL, Reason: err.Error()}
	}

	// Normalize the path
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	// Sanitize path encoding
	parsed.RawPath = ""
	parsed.Path = sanitizePath(parsed.Path)

	return parsed, nil
}

// validateHostname validates hostname/domain format and constraints
func validateHostname(host string) error {
	// Remove port if present
	hostname := host
	if idx := strings.Index(host, ":"); idx > 0 {
		hostname = host[:idx]
	}

	// Basic length check
	if len(hostname) == 0 {
		return fmt.Errorf("empty hostname")
	}
	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}

	// Check for IPv4 addresses
	if isIPv4Address(hostname) {
		return nil // IPv4 is valid
	}

	// Check for IPv6 addresses
	if strings.HasPrefix(hostname, "[") && strings.HasSuffix(hostname, "]") {
		if isIPv6Address(hostname[1 : len(hostname)-1]) {
			return nil // IPv6 is valid
		}
	}

	// Validate domain name format
	if err := validateDomainName(hostname); err != nil {
		return fmt.Errorf("invalid domain name: %v", err)
	}

	return nil
}

// isIPv4Address checks if string is a valid IPv4 address
func isIPv4Address(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		num := 0
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
			num = num*10 + int(r-'0')
		}
		if num > 255 {
			return false
		}
	}
	return true
}

// isIPv6Address checks if string is a valid IPv6 address
func isIPv6Address(s string) bool {
	// Basic IPv6 validation - simplified check
	if len(s) == 0 {
		return false
	}

	// Count colons
	colonCount := strings.Count(s, ":")
	if colonCount < 2 || colonCount > 7 {
		return false
	}

	// Basic regex pattern for IPv6
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|::([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{1,4}$`)
	return ipv6Regex.MatchString(s)
}

// validateDomainName validates domain name format
func validateDomainName(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("empty domain")
	}

	// Domain name regex - simplified version
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain name format")
	}

	// Check for consecutive hyphens or dots
	if strings.Contains(domain, "--") || strings.Contains(domain, "..") {
		return fmt.Errorf("consecutive hyphens or dots not allowed")
	}

	return nil
}

// sanitizePath sanitizes URL path for security and consistency
func sanitizePath(path string) string {
	// Decode and re-encode to normalize encoding
	if decodedPath, err := url.QueryUnescape(path); err == nil {
		path = url.QueryEscape(decodedPath)
		path = strings.ReplaceAll(path, "%2F", "/") // Keep slashes as-is
		path = strings.ReplaceAll(path, "%2f", "/") // Handle lowercase too
	}

	// Remove dangerous sequences
	path = strings.ReplaceAll(path, "../", "")  // Directory traversal
	path = strings.ReplaceAll(path, "..\\", "") // Windows directory traversal
	path = strings.ReplaceAll(path, "./", "")   // Current directory references

	return path
}

// FormatJavaScriptURL formats relative JavaScript URLs into absolute URLs
func FormatJavaScriptURL(jsSrc string, baseURL *url.URL) (string, error) {
	if jsSrc == "" {
		return "", nil
	}

	// If it's already absolute, return as-is
	if strings.HasPrefix(jsSrc, "http://") || strings.HasPrefix(jsSrc, "https://") {
		// Quick validation
		if _, err := url.Parse(jsSrc); err != nil {
			return "", URLValidationError{URL: jsSrc, Reason: fmt.Sprintf("invalid JavaScript URL: %v", err)}
		}
		return jsSrc, nil
	}

	var jsURL string
	if strings.HasPrefix(jsSrc, "//") {
		// Protocol-relative URL
		jsURL = fmt.Sprintf("%s:%s", baseURL.Scheme, jsSrc)
	} else if strings.HasPrefix(jsSrc, "/") {
		// Absolute path
		jsURL = fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, jsSrc)
	} else {
		// Relative path - resolve against base
		if baseURL.Path == "" || baseURL.Path == "/" {
			jsURL = fmt.Sprintf("%s://%s/%s", baseURL.Scheme, baseURL.Host, jsSrc)
		} else {
			basePath := baseURL.Path
			if !strings.HasSuffix(basePath, "/") {
				basePath = basePath[:strings.LastIndex(basePath, "/")+1]
			}
			jsURL = fmt.Sprintf("%s://%s%s%s", baseURL.Scheme, baseURL.Host, basePath, jsSrc)
		}
	}

	return jsURL, nil
}
