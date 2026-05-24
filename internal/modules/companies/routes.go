package companies

import (
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

// Register mounts companies module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc, requireRole func(string) gin.HandlerFunc) {
	companies := v1.Group("/companies")
	{
		companies.GET("", requirePermission(platformPerms.Format(platformPerms.ModuleCompanies, platformPerms.ActionList)), handler.List)
		companies.GET("/:id", requirePermission(platformPerms.Format(platformPerms.ModuleCompanies, platformPerms.ActionView)), handler.GetByPublicID)
		companies.PUT("/:id", requireRole("root"), handler.Update)
		companies.DELETE("/:id", requireRole("root"), handler.Delete)
	}
}
