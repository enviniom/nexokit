package permissions

import "github.com/gin-gonic/gin"

// Register mounts permissions module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	permissions := v1.Group("/permissions")
	{
		permissions.GET("", requirePermission("permissions.manage"), handler.List)
		permissions.GET("/:id", requirePermission("permissions.manage"), handler.GetByPublicID)
		permissions.POST("", requirePermission("permissions.manage"), handler.Create)
		permissions.PUT("/:id", requirePermission("permissions.manage"), handler.Update)
		permissions.DELETE("/:id", requirePermission("permissions.manage"), handler.Delete)
	}
}
