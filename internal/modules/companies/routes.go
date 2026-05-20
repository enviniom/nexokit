package companies

import "github.com/gin-gonic/gin"

// Register mounts companies module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc, requireRole func(string) gin.HandlerFunc) {
	companies := v1.Group("/companies")
	{
		companies.GET("", requirePermission("companies.index"), handler.List)
		companies.GET("/:id", requirePermission("companies.view"), handler.GetByPublicID)
		companies.POST("", requireRole("root"), handler.Create)
		companies.PUT("/:id", requireRole("root"), handler.Update)
		companies.DELETE("/:id", requireRole("root"), handler.Delete)
	}
}
