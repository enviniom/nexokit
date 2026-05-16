package users

import "github.com/gin-gonic/gin"

// Register mounts users module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler) {
	users := v1.Group("/users")
	{
		users.GET("", handler.List)
		users.POST("", handler.Create)
		users.GET("/:id", handler.GetByPublicID)
		users.PUT("/:id", handler.Update)
		users.DELETE("/:id", handler.Delete)
		users.PATCH("/:id/password", handler.ChangePassword)
		users.PATCH("/:id/status", handler.ToggleStatus)
	}
}
