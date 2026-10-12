package slices

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/internal"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/internal/list_all_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/internal/resolve_auth_user"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/internal/resolve_role_by_slug"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/internal/resolve_user_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/internal/sync_permissions"
	"gorm.io/gorm"
)

// Capabilities exposes module contracts without exposing internal slice implementations.
type Capabilities struct {
	AuthUserResolver core.AuthUserResolver
	Resolver         core.PermissionResolver
	Syncer           core.PermissionSyncer
	RoleResolver     core.RoleBySlugResolver
	Catalog          core.PermissionCatalogReader
}

// NewCapabilities assembles internal slices within their Go visibility boundary.
func NewCapabilities(db *gorm.DB, c cache.Cache) *Capabilities {
	return &Capabilities{
		AuthUserResolver: internal.NewResolveAuthUserService(resolve_auth_user.NewRepository(db)),
		Resolver:         internal.NewResolveUserPermissionsService(resolve_user_permissions.NewRepository(db), c),
		Syncer:           internal.NewSyncPermissionsService(sync_permissions.NewRepository(db)),
		RoleResolver:     internal.NewResolveRoleBySlugService(resolve_role_by_slug.NewRepository(db)),
		Catalog:          internal.NewListAllPermissionsService(list_all_permissions.NewRepository(db)),
	}
}
