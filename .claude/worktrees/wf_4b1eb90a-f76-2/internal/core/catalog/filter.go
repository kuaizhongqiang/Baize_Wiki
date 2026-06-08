package catalog

import (
	"regexp"
)

// Patterns for sensitive information that should be redacted from summaries.
var sensitivePatterns = []*regexp.Regexp{
	// AK/SK/key/secret/token/password with values (quoted or bare)
	regexp.MustCompile(`(?i)(?:AK|SK|secret|token|password|api.?key|access.?key)\s*[:=]\s*["']?[A-Za-z0-9_\-]{8,}["']?`),
	// Bearer tokens
	regexp.MustCompile(`Bearer\s+[A-Za-z0-9_\-\.]{20,}`),
	// Private key markers
	regexp.MustCompile(`-----BEGIN\s+(RSA|EC|DSA|OPENSSH)\s+PRIVATE\s+KEY-----`),
	// Internal IP addresses
	regexp.MustCompile(`10\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
	regexp.MustCompile(`172\.(1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}`),
	regexp.MustCompile(`192\.168\.\d{1,3}\.\d{1,3}`),
	// Connection strings
	regexp.MustCompile(`(?:jdbc|postgres|mysql|mongodb|redis)://[^\s"]+`),
}

// FilterSensitive replaces detected sensitive information with [REDACTED].
func FilterSensitive(input string) string {
	if input == "" {
		return input
	}
	result := input
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}
