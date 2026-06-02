package iam

import (
	"github.com/enviniom/nexokit/internal/modules/iam/permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/roles"
	"github.com/enviniom/nexokit/internal/modules/iam/users"
	"github.com/gin-gonic/gin"
)

// Register mounts IAM routes onto v1 router groups.
// Actual entity route wiring is completed in later PR slices.
func Register(globalProtected *gin.RouterGroup, c *Container, tenantProtected *gin.RouterGroup, requirePermission func(string) gin.HandlerFunc, _ func(string) gin.HandlerFunc) {
	users.Register(tenantProtected, c.Users, requirePermission)
	roles.Register(globalProtected, c.Roles, requirePermission)
	permissions.Register(globalProtected, c.Permissions, requirePermission)
}
