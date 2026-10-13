// Package roles owns application role identities and reserved-identity policy.
package roles

import "strings"

const (
	RootRoleSlug  = "root"
	AdminRoleSlug = "admin"
	UserRoleSlug  = "user"
)

var ReservedSlugs = []string{RootRoleSlug, AdminRoleSlug, UserRoleSlug}

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
