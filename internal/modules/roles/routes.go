package roles

import "github.com/gin-gonic/gin"

// Register mounts roles module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	roles := v1.Group("/roles")
	{
		roles.GET("", requirePermission("roles.index"), handler.List)
		roles.GET("/:id", requirePermission("roles.view"), handler.GetByPublicID)
		roles.GET("/:id/permissions", requirePermission("roles.view"), handler.GetPermissionCatalog)
		roles.PUT("/:id/permissions", requirePermission("roles.assign_permissions"), handler.AssignPermissions)
	}
}
