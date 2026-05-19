package auth

import "github.com/gin-gonic/gin"

// Register mounts auth module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc, attachPermissions gin.HandlerFunc) {
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
		auth.GET("/me", authMiddleware, attachPermissions, handler.Me)
	}
}
