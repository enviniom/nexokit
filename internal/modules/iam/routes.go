package iam

import (
	"github.com/enviniom/nexokit/internal/modules/iam/slices/permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/roles"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/users"
	"github.com/gin-gonic/gin"
)

// Register mounts IAM routes onto v1 router groups.
func Register(globalProtected *gin.RouterGroup, c *Container, tenantProtected *gin.RouterGroup, requirePermission func(string) gin.HandlerFunc, _ func(string) gin.HandlerFunc) {
	users.Register(tenantProtected, c.Users, requirePermission)
	roles.Register(globalProtected, c.Roles, requirePermission)
	permissions.Register(globalProtected, c.Permissions, requirePermission)
}
