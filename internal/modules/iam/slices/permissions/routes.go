package permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc) {
	permissions := v1.Group("/permissions")
	{
		permissions.GET("", requirePermission(platformPerms.Format(core.ModulePermissions, platformPerms.ActionManage)), container.ListHandler.Handle)
		permissions.GET("/:id", requirePermission(platformPerms.Format(core.ModulePermissions, platformPerms.ActionManage)), container.ViewHandler.Handle)
		permissions.PUT("/:id", requirePermission(platformPerms.Format(core.ModulePermissions, platformPerms.ActionManage)), container.UpdateHandler.Handle)
	}
}
