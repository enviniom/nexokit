package roles

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/assign_permissions_to_role"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/create_role"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/delete_role"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/list_roles"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/list_selectable_roles"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/update_role"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/view_role"
	"github.com/enviniom/nexokit/internal/modules/iam/roles/view_role_permission_catalog"
	"gorm.io/gorm"
)

type Container struct {
	ListHandler                    *list_roles.Handler
	ListSelectableHandler          *list_selectable_roles.Handler
	ViewHandler                    *view_role.Handler
	CreateHandler                  *create_role.Handler
	UpdateHandler                  *update_role.Handler
	DeleteHandler                  *delete_role.Handler
	ViewPermissionCatalogHandler   *view_role_permission_catalog.Handler
	AssignPermissionsToRoleHandler *assign_permissions_to_role.Handler
}

func NewContainer(db *gorm.DB, c cache.Cache) *Container {
	viewRoleRepo := view_role.NewRepository(db)
	listRolesRepo := list_roles.NewRepository(db)
	createRoleRepo := create_role.NewRepository(db)
	updateRoleRepo := update_role.NewRepository(db)
	deleteRoleRepo := delete_role.NewRepository(db)
	assignPermissionsToRoleRepo := assign_permissions_to_role.NewRepository(db)
	listSelectableRolesRepo := list_selectable_roles.NewRepository(db)
	viewRolePermissionCatalogRepo := view_role_permission_catalog.NewRepository(db)

	return &Container{
		ListHandler:                    list_roles.NewHandler(list_roles.NewService(listRolesRepo)),
		ListSelectableHandler:          list_selectable_roles.NewHandler(list_selectable_roles.NewService(listSelectableRolesRepo)),
		ViewHandler:                    view_role.NewHandler(view_role.NewService(viewRoleRepo)),
		CreateHandler:                  create_role.NewHandler(create_role.NewService(createRoleRepo)),
		UpdateHandler:                  update_role.NewHandler(update_role.NewService(updateRoleRepo)),
		DeleteHandler:                  delete_role.NewHandler(delete_role.NewService(deleteRoleRepo)),
		ViewPermissionCatalogHandler:   view_role_permission_catalog.NewHandler(view_role_permission_catalog.NewService(viewRolePermissionCatalogRepo)),
		AssignPermissionsToRoleHandler: assign_permissions_to_role.NewHandler(assign_permissions_to_role.NewService(assignPermissionsToRoleRepo, c)),
	}
}
