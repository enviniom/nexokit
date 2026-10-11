package goldenmod

import "github.com/gin-gonic/gin"

// Register mounts prepared handlers. The app must explicitly call this on a
// group with authentication, authorization and (for --tenant) tenant middleware.
// Only List accepts explicit trusted root-global scope; other tenant operations
// require a selected company. Nothing is automatically mounted by generation.
func Register(v1 *gin.RouterGroup, container *Container) {
	g := v1.Group("/goldenmod")
	g.POST("", container.Create.Handle)
	g.GET("", container.List.Handle)
	g.GET("/:id", container.Get.Handle)
	g.PUT("/:id", container.Update.Handle)
	g.DELETE("/:id", container.Delete.Handle)
}
