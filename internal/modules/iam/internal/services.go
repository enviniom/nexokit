package internal

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/list_all_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_auth_user"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_role_by_slug"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_user_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/sync_permissions"
)

// NewResolveAuthUserService wires the resolve_auth_user slice using the supplied repository.
func NewResolveAuthUserService(repo resolve_auth_user.Repository) core.AuthUserResolver {
	return resolve_auth_user.NewService(repo)
}

// NewResolveUserPermissionsService wires the resolve_user_permissions slice using the supplied repository and cache.
func NewResolveUserPermissionsService(repo resolve_user_permissions.Repository, c cache.Cache) core.PermissionResolver {
	return resolve_user_permissions.NewService(repo, c)
}

// NewSyncPermissionsService wires the sync_permissions slice using the supplied repository.
func NewSyncPermissionsService(repo sync_permissions.Repository) core.PermissionSyncer {
	return sync_permissions.NewService(repo)
}

// NewResolveRoleBySlugService wires the resolve_role_by_slug slice using the supplied repository.
func NewResolveRoleBySlugService(repo resolve_role_by_slug.Repository) core.RoleBySlugResolver {
	return resolve_role_by_slug.NewService(repo)
}

// NewListAllPermissionsService wires the list_all_permissions slice using the supplied repository.
func NewListAllPermissionsService(repo list_all_permissions.Repository) core.PermissionCatalogReader {
	return list_all_permissions.NewService(repo)
}
