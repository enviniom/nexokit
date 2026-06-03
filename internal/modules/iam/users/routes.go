package users

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc) {
	users := v1.Group("/users")
	{
		users.GET("", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionList)), container.ListHandler.Handle)
		users.POST("", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionCreate)), container.CreateHandler.Handle)
		users.GET("/:id", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionView)), container.ViewHandler.Handle)
		users.PUT("/:id", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionUpdate)), container.UpdateHandler.Handle)
		users.DELETE("/:id", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionDelete)), container.DeleteHandler.Handle)
		users.PATCH("/:id/password", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionUpdate)), container.ChangePasswordHandler.Handle)
		users.PATCH("/:id/role", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionChangeRole)), container.AssignRoleHandler.Handle)
		users.PATCH("/:id/status", requirePermission(platformPerms.Format(core.ModuleUsers, platformPerms.ActionUpdate)), container.ToggleStatusHandler.Handle)
	}
}
