package onboarding

import "github.com/gin-gonic/gin"

// Register mounts onboarding module routes onto the v1 global protected router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requireRole func(string) gin.HandlerFunc) {
	onboarding := v1.Group("/onboarding")
	{
		onboarding.POST("/companies", requireRole("root"), handler.OnboardCompany)
	}
}
