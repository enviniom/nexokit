package permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/permissions/list_permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/permissions/update_permission"
	"github.com/enviniom/nexokit/internal/modules/iam/permissions/view_permission"
	"gorm.io/gorm"
)

type Container struct {
	ListHandler   *list_permissions.Handler
	ViewHandler   *view_permission.Handler
	UpdateHandler *update_permission.Handler
}

func NewContainer(db *gorm.DB) *Container {
	listRepo := list_permissions.NewRepository(db)
	viewRepo := view_permission.NewRepository(db)
	updateRepo := update_permission.NewRepository(db)
	return &Container{
		ListHandler:   list_permissions.NewHandler(list_permissions.NewService(listRepo)),
		ViewHandler:   view_permission.NewHandler(view_permission.NewService(viewRepo)),
		UpdateHandler: update_permission.NewHandler(update_permission.NewService(updateRepo)),
	}
}
