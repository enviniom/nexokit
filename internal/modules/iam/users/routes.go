package users

import (
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc) {
	users := v1.Group("/users")
	{
		users.GET("", requirePermission(platformPerms.Format("users", platformPerms.ActionList)), container.ListHandler.Handle)
		users.POST("", requirePermission(platformPerms.Format("users", platformPerms.ActionCreate)), container.CreateHandler.Handle)
		users.GET("/:id", requirePermission(platformPerms.Format("users", platformPerms.ActionView)), container.ViewHandler.Handle)
		users.PUT("/:id", requirePermission(platformPerms.Format("users", platformPerms.ActionUpdate)), container.UpdateHandler.Handle)
		users.DELETE("/:id", requirePermission(platformPerms.Format("users", platformPerms.ActionDelete)), container.DeleteHandler.Handle)
		users.PATCH("/:id/password", requirePermission(platformPerms.Format("users", platformPerms.ActionUpdate)), container.ChangePasswordHandler.Handle)
		users.PATCH("/:id/role", requirePermission(platformPerms.Format("users", platformPerms.ActionChangeRole)), container.AssignRoleHandler.Handle)
		users.PATCH("/:id/status", requirePermission(platformPerms.Format("users", platformPerms.ActionUpdate)), container.ToggleStatusHandler.Handle)
	}
}
