package permissions

import (
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

// Register mounts permissions module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	permissions := v1.Group("/permissions")
	{
		permissions.GET("", requirePermission(platformPerms.Format(platformPerms.ModulePermissions, platformPerms.ActionManage)), handler.List)
		permissions.GET("/:id", requirePermission(platformPerms.Format(platformPerms.ModulePermissions, platformPerms.ActionManage)), handler.GetByPublicID)
		permissions.PUT("/:id", requirePermission(platformPerms.Format(platformPerms.ModulePermissions, platformPerms.ActionManage)), handler.Update)
	}
}
