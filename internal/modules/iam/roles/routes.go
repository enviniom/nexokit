package roles

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc) {
	roles := v1.Group("/roles")
	{
		roles.GET("", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionList)), container.ListHandler.Handle)
		roles.GET("/select", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionSelect)), container.ListSelectableHandler.Handle)
		roles.GET("/:id", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionView)), container.ViewHandler.Handle)
		roles.POST("", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionCreate)), container.CreateHandler.Handle)
		roles.PUT("/:id", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionUpdate)), container.UpdateHandler.Handle)
		roles.DELETE("/:id", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionDelete)), container.DeleteHandler.Handle)
		roles.GET("/:id/permissions", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionView)), container.ViewPermissionCatalogHandler.Handle)
		roles.PUT("/:id/permissions", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionAssignPermissions)), container.AssignPermissionsToRoleHandler.Handle)
	}
}
