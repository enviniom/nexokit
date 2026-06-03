package core

const (
	ModuleIAM         = "iam"
	ModuleUsers       = "users"
	ModuleRoles       = "roles"
	ModulePermissions = "permissions"

	RootRoleSlug  = "root"
	AdminRoleSlug = "admin"
	UserRoleSlug  = "user"
)

var ReservedSlugs = []string{RootRoleSlug, AdminRoleSlug, UserRoleSlug}
