package goldenmod

import "github.com/gin-gonic/gin"

// Register mounts Goldenmod routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, h *GoldenmodHandler) {
	g := v1.Group("/goldenmod")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
