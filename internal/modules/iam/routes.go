package iam

import (
	"github.com/enviniom/nexokit/internal/modules/iam/slices/auth"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/roles"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/users"
	"github.com/gin-gonic/gin"
)

// Register mounts IAM routes onto v1 router groups.
func Register(v1, globalProtected, tenantProtected *gin.RouterGroup, c *Container, requirePermission func(string) gin.HandlerFunc, authMiddleware, loginRateLimit, refreshRateLimit gin.HandlerFunc) {
	auth.Register(v1, c.Auth, authMiddleware, loginRateLimit, refreshRateLimit)
	users.Register(tenantProtected, c.Users, requirePermission)
	roles.Register(globalProtected, c.Roles, requirePermission)
	permissions.Register(globalProtected, c.Permissions, requirePermission)
}
