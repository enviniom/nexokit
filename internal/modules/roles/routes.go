package roles

import "github.com/gin-gonic/gin"

// Register mounts roles module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler) {
	roles := v1.Group("/roles")
	{
		roles.GET("", handler.List)
		roles.POST("", handler.Create)
		roles.GET("/:id", handler.GetByPublicID)
		roles.PUT("/:id", handler.Update)
		roles.DELETE("/:id", handler.Delete)
	}
}
