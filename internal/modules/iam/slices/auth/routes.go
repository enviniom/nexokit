package auth

import "github.com/gin-gonic/gin"

// Register mounts auth module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, container *Container, authMiddleware gin.HandlerFunc, loginRateLimit gin.HandlerFunc, refreshRateLimit gin.HandlerFunc) {
	auth := v1.Group("/auth")
	{
		auth.POST("/login", loginRateLimit, container.AuthenticateUser.Handle)
		auth.POST("/refresh", refreshRateLimit, container.RotateToken.Handle)
		auth.POST("/logout", container.RevokeToken.Handle)
		auth.GET("/me", authMiddleware, container.ViewSession.Handle)
	}
}
