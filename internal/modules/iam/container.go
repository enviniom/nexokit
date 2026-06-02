package iam

import (
	"log/slog"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iaminternal "github.com/enviniom/nexokit/internal/modules/iam/internal"
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
		Users:           users.NewContainer(db, c),
		Roles:           roles.NewContainer(db, c),
		Permissions:     permissions.NewContainer(db),
		AuthUserResolver: iaminternal.NewResolveAuthUserService(db),
		Resolver:         iaminternal.NewResolveUserPermissionsService(db, c),
		Syncer:           iaminternal.NewSyncPermissionsService(db),
		RoleResolver:     iaminternal.NewResolveRoleBySlugService(db),
		Catalog:          iaminternal.NewListAllPermissionsService(db),
	}
}
