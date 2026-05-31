package companies

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
)

// Register mounts companies module routes onto the v1 router group.
func Register(v1 *gin.RouterGroup, container *Container, requirePermission func(string) gin.HandlerFunc, requireRole func(string) gin.HandlerFunc) {
	companies := v1.Group("/companies")
	{
		companies.GET("", requireRole("root"), container.ListCompanies.Handle)
		companies.GET("/:id", requirePermission(platformPerms.Format(core.ModuleCompanies, platformPerms.ActionView)), container.ViewCompany.Handle)
		companies.PUT("/:id", requireRole("root"), container.UpdateCompany.Handle)
		companies.DELETE("/:id", requireRole("root"), container.DeleteCompany.Handle)
		companies.GET("/:id/domains", requireRole("root"), container.ListCompanyDomains.Handle)
		companies.POST("/:id/domains", requireRole("root"), container.CreateCompanyDomain.Handle)
		companies.PUT("/:id/domains/:domain_id", requireRole("root"), container.UpdateCompanyDomain.Handle)
	}
}
