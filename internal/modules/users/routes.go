package users

import (
	"github.com/enviniom/nexokit/internal/modules/users/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

// Register mounts users module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	users := v1.Group("/users")
	{
		users.GET("", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionList)), handler.List)
		users.POST("", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionCreate)), handler.Create)
		users.GET("/:id", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionView)), handler.GetByPublicID)
		users.PUT("/:id", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionUpdate)), handler.Update)
		users.PATCH("/:id/role", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionChangeRole)), handler.ChangeRole)
		users.DELETE("/:id", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionDelete)), handler.Delete)
		users.PATCH("/:id/password", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionUpdate)), handler.ChangePassword)
		users.PATCH("/:id/status", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionUpdate)), handler.ToggleStatus)
	}
}
