package str

import "strings"

// NormalizeDomain returns a trimmed, lower-case version of s with a single
// trailing dot removed.
func NormalizeDomain(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.TrimSuffix(s, ".")
	return s
}
