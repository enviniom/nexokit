package users

import "github.com/gin-gonic/gin"

// Register mounts users module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	users := v1.Group("/users")
	{
		users.GET("", requirePermission("users.index"), handler.List)
		users.POST("", requirePermission("users.create"), handler.Create)
		users.GET("/:id", requirePermission("users.view"), handler.GetByPublicID)
		users.PUT("/:id", requirePermission("users.update"), requirePermission("users.change_role"), handler.Update)
		users.DELETE("/:id", requirePermission("users.delete"), handler.Delete)
		users.PATCH("/:id/password", requirePermission("users.update"), handler.ChangePassword)
		users.PATCH("/:id/status", requirePermission("users.update"), handler.ToggleStatus)
	}
}
