package permissions

import "github.com/gin-gonic/gin"

// Register mounts permissions module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler) {
	permissions := v1.Group("/permissions")
	{
		permissions.GET("", handler.List)
		permissions.GET("/:id", handler.GetByPublicID)
		permissions.POST("", handler.Create)
		permissions.PUT("/:id", handler.Update)
		permissions.DELETE("/:id", handler.Delete)
	}
}
