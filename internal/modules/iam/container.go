package iam

import (
	"log/slog"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iaminternal "github.com/enviniom/nexokit/internal/modules/iam/internal"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/list_all_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_auth_user"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_role_by_slug"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/resolve_user_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/internal/sync_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/roles"
	"github.com/enviniom/nexokit/internal/modules/iam/users"
	"gorm.io/gorm"
)

// Container is the root IAM container used by app wiring.
// Foundation slice exposes contracts first; entity sub-containers are wired in later slices.
type Container struct {
	Users       *users.Container
	Roles       *roles.Container
	Permissions *permissions.Container

	AuthUserResolver core.AuthUserResolver
	Resolver         core.PermissionResolver
	Syncer           core.PermissionSyncer
	RoleResolver     core.RoleBySlugResolver
	Catalog          core.PermissionCatalogReader
}

func NewContainer(db *gorm.DB, c cache.Cache, _ *slog.Logger) *Container {
	return &Container{
		Users:            users.NewContainer(db, c),
		Roles:            roles.NewContainer(db, c),
		Permissions:      permissions.NewContainer(db),
		AuthUserResolver: iaminternal.NewResolveAuthUserService(resolve_auth_user.NewRepository(db)),
		Resolver:         iaminternal.NewResolveUserPermissionsService(resolve_user_permissions.NewRepository(db), c),
		Syncer:           iaminternal.NewSyncPermissionsService(sync_permissions.NewRepository(db)),
		RoleResolver:     iaminternal.NewResolveRoleBySlugService(resolve_role_by_slug.NewRepository(db)),
		Catalog:          iaminternal.NewListAllPermissionsService(list_all_permissions.NewRepository(db)),
	}
}
