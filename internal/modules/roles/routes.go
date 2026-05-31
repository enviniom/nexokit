package roles

import (
	"github.com/enviniom/nexokit/internal/modules/roles/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

// Register mounts roles module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	roles := v1.Group("/roles")
	{
		roles.GET("", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionList)), handler.List)
		roles.GET("/select", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionSelect)), handler.ListSelect)
		roles.GET("/:id", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionView)), handler.GetByPublicID)
		roles.POST("", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionCreate)), handler.Create)
		roles.PUT("/:id", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionUpdate)), handler.Update)
		roles.DELETE("/:id", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionDelete)), handler.Delete)
		roles.GET("/:id/permissions", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionView)), handler.GetPermissionCatalog)
		roles.PUT("/:id/permissions", requirePermission(platformPerms.Format(core.ModuleRoles, platformPerms.ActionAssignPermissions)), handler.AssignPermissions)
	}
}
