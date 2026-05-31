package permissions

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

// Register mounts permissions module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc) {
	permissions := v1.Group("/permissions")
	{
		permissions.GET("", requirePermission(platformPerms.Format(core.ModulePermissions, platformPerms.ActionManage)), container.ListHandler.List)
		permissions.GET("/:id", requirePermission(platformPerms.Format(core.ModulePermissions, platformPerms.ActionManage)), container.ViewHandler.GetByPublicID)
		permissions.PUT("/:id", requirePermission(platformPerms.Format(core.ModulePermissions, platformPerms.ActionManage)), container.UpdateHandler.Update)
	}
}
