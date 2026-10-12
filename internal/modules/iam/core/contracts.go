package core

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/authctx"
)

// AuthUserResolver resolves a user for authentication middleware.
type AuthUserResolver interface {
	ResolveAuthUser(publicID string) (*authctx.User, error)
}

// PermissionResolver resolves permission slugs for a user public ID.
type PermissionResolver interface {
	Resolve(publicID string) ([]string, error)
}

// PermissionSyncer synchronizes registered permissions.
type PermissionSyncer interface {
	SyncPermissions(slugs []string) error
}

// RoleBySlugResolver resolves a role by slug.
type RoleBySlugResolver interface {
	ResolveRoleBySlug(slug string) (*iamrole.IAMRole, error)
}

// PermissionCatalogReader returns all permissions.
type PermissionCatalogReader interface {
	ListAllPermissions() ([]permission.IAMPermission, error)
}
