package permissions

import (
	"log/slog"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/modules/permissions/list_permissions"
	"github.com/enviniom/nexokit/internal/modules/permissions/resolve_permissions"
	"github.com/enviniom/nexokit/internal/modules/permissions/sync_permissions"
	"github.com/enviniom/nexokit/internal/modules/permissions/update_permission"
	"github.com/enviniom/nexokit/internal/modules/permissions/view_permission"
	"gorm.io/gorm"
)

type Container struct {
	ListHandler   *list_permissions.Handler
	ViewHandler   *view_permission.Handler
	UpdateHandler *update_permission.Handler

	Resolver core.Resolver
	Syncer   core.Syncer
	Catalog  core.PermissionCatalogReader
}

func NewContainer(db *gorm.DB, c cache.Cache, _ *slog.Logger) *Container {
	listRepo := list_permissions.NewRepository(db)
	listService := list_permissions.NewService(listRepo)

	viewRepo := view_permission.NewRepository(db)
	viewService := view_permission.NewService(viewRepo)

	updateRepo := update_permission.NewRepository(db)
	updateService := update_permission.NewService(updateRepo)

	resolveRepo := resolve_permissions.NewRepository(db)
	resolver := resolve_permissions.NewService(resolveRepo, c)

	syncRepo := sync_permissions.NewRepository(db)
	syncer := sync_permissions.NewService(syncRepo)

	return &Container{
		ListHandler:   list_permissions.NewHandler(listService),
		ViewHandler:   view_permission.NewHandler(viewService),
		UpdateHandler: update_permission.NewHandler(updateService),
		Resolver:      resolver,
		Syncer:        syncer,
		Catalog:       listRepo,
	}
}
