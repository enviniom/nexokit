package str

import "strings"

// NormalizeEmail returns a trimmed, lower-case version of s.
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
