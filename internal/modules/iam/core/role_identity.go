package core

import "strings"

func IsReservedRoleIdentity(name, slug string) bool {
	n := strings.TrimSpace(name)
	s := strings.TrimSpace(slug)
	for _, reserved := range ReservedSlugs {
		if strings.EqualFold(n, reserved) || strings.EqualFold(s, reserved) {
			return true
		}
	}
	return false
}
