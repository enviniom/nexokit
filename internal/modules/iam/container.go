package iam

import (
	"log/slog"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/slices"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/roles"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/users"
	"gorm.io/gorm"
)

// Container is the root IAM container used by app wiring.
// Entity containers and internal capabilities share the supplied database and cache.
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
	usersContainer := users.NewContainer(db, c)
	rolesContainer := roles.NewContainer(db, c)
	permissionsContainer := permissions.NewContainer(db)
	capabilities := slices.NewCapabilities(db, c)

	return &Container{
		Users:            usersContainer,
		Roles:            rolesContainer,
		Permissions:      permissionsContainer,
		AuthUserResolver: capabilities.AuthUserResolver,
		Resolver:         capabilities.Resolver,
		Syncer:           capabilities.Syncer,
		RoleResolver:     capabilities.RoleResolver,
		Catalog:          capabilities.Catalog,
	}
}
