package internal

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/list_all_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_auth_user"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_role_by_slug"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_user_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/sync_permissions"
	"gorm.io/gorm"
)

func NewResolveAuthUserService(db *gorm.DB) core.AuthUserResolver {
	return resolve_auth_user.NewService(resolve_auth_user.NewRepository(db))
}

func NewResolveUserPermissionsService(db *gorm.DB, c cache.Cache) core.PermissionResolver {
	return resolve_user_permissions.NewService(resolve_user_permissions.NewRepository(db), c)
}

func NewSyncPermissionsService(db *gorm.DB) core.PermissionSyncer {
	return sync_permissions.NewService(sync_permissions.NewRepository(db))
}

func NewResolveRoleBySlugService(db *gorm.DB) core.RoleBySlugResolver {
	return resolve_role_by_slug.NewService(resolve_role_by_slug.NewRepository(db))
}

func NewListAllPermissionsService(db *gorm.DB) core.PermissionCatalogReader {
	return list_all_permissions.NewService(list_all_permissions.NewRepository(db))
}
