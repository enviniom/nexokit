package roles

import (
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc) {
	roles := v1.Group("/roles")
	{
		roles.GET("", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionList)), container.ListHandler.Handle)
		roles.GET("/select", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionSelect)), container.ListSelectableHandler.Handle)
		roles.GET("/:id", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionView)), container.ViewHandler.Handle)
		roles.POST("", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionCreate)), container.CreateHandler.Handle)
		roles.PUT("/:id", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionUpdate)), container.UpdateHandler.Handle)
		roles.DELETE("/:id", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionDelete)), container.DeleteHandler.Handle)
		roles.GET("/:id/permissions", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionView)), container.ViewPermissionCatalogHandler.Handle)
		roles.PUT("/:id/permissions", requirePermission(platformPerms.Format(platformPerms.ModuleRoles, platformPerms.ActionAssignPermissions)), container.AssignPermissionsToRoleHandler.Handle)
	}
}
