package core

const (
	ModuleIAM = "iam"

	RootRoleSlug  = "root"
	AdminRoleSlug = "admin"
	UserRoleSlug  = "user"
)

var ReservedSlugs = []string{RootRoleSlug, AdminRoleSlug, UserRoleSlug}
