package companies

import (
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	platformRoles "github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/gin-gonic/gin"
)

// Register mounts companies module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc, requireRole func(string) gin.HandlerFunc) {
	companies := v1.Group("/companies")
	{
		companies.GET("", requireRole(platformRoles.RootRoleSlug), container.ListCompanies.Handle)
		companies.GET("/:id", requirePermission(platformPerms.Format(platformPerms.ModuleCompanies, platformPerms.ActionView)), container.ViewCompany.Handle)
		companies.PUT("/:id", requireRole(platformRoles.RootRoleSlug), container.UpdateCompany.Handle)
		companies.DELETE("/:id", requireRole(platformRoles.RootRoleSlug), container.DeleteCompany.Handle)
		companies.GET("/:id/domains", requireRole(platformRoles.RootRoleSlug), container.ListCompanyDomains.Handle)
		companies.POST("/:id/domains", requireRole(platformRoles.RootRoleSlug), container.CreateCompanyDomain.Handle)
		companies.PUT("/:id/domains/:domain_id", requireRole(platformRoles.RootRoleSlug), container.UpdateCompanyDomain.Handle)
	}
}
