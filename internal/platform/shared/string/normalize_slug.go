package str

import "strings"

// NormalizeSlug returns a trimmed, lower-case version of s.
func NormalizeSlug(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
