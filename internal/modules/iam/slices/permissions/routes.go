package permissions

import (
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	platformRoles "github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/gin-gonic/gin"
)

func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc, requireRole func(string) gin.HandlerFunc) {
	permissions := v1.Group("/permissions")
	{
		permissions.GET("", requirePermission(platformPerms.Format(platformPerms.ModulePermissions, platformPerms.ActionManage)), container.ListHandler.Handle)
		permissions.GET("/:id", requirePermission(platformPerms.Format(platformPerms.ModulePermissions, platformPerms.ActionManage)), container.ViewHandler.Handle)
		permissions.PUT("/:id", requireRole(platformRoles.RootRoleSlug), container.UpdateHandler.Handle)
	}
}
