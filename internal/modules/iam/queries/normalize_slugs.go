package queries

import "strings"

func NormalizeSlugs(slugs []string) []string {
	seen, out := map[string]bool{}, make([]string, 0, len(slugs))
	for _, s := range slugs {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
