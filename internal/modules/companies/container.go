package companies

import (
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/companies/create_company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/delete_company"
	"github.com/enviniom/nexokit/internal/modules/companies/list_companies"
	"github.com/enviniom/nexokit/internal/modules/companies/list_company_domains"
	"github.com/enviniom/nexokit/internal/modules/companies/update_company"
	"github.com/enviniom/nexokit/internal/modules/companies/update_company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/view_company"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Container struct {
	ListCompanies       *list_companies.Handler
	ViewCompany         *view_company.Handler
	UpdateCompany       *update_company.Handler
	DeleteCompany       *delete_company.Handler
	ListCompanyDomains  *list_company_domains.Handler
	CreateCompanyDomain *create_company_domain.Handler
	UpdateCompanyDomain *update_company_domain.Handler
	resolver            middleware.CompanyResolver
}

func NewContainer(db *gorm.DB) *Container {
	listRepo := list_companies.NewRepository(db)
	viewRepo := view_company.NewRepository(db)
	updateRepo := update_company.NewRepository(db)
	deleteRepo := delete_company.NewRepository(db)
	listDomainRepo := list_company_domains.NewRepository(db)
	createDomainRepo := create_company_domain.NewRepository(db)
	updateDomainRepo := update_company_domain.NewRepository(db)

	return &Container{
		ListCompanies:       list_companies.NewHandler(list_companies.NewService(listRepo)),
		ViewCompany:         view_company.NewHandler(view_company.NewService(viewRepo)),
		UpdateCompany:       update_company.NewHandler(update_company.NewService(updateRepo)),
		DeleteCompany:       delete_company.NewHandler(delete_company.NewService(deleteRepo)),
		ListCompanyDomains:  list_company_domains.NewHandler(list_company_domains.NewService(listDomainRepo)),
		CreateCompanyDomain: create_company_domain.NewHandler(create_company_domain.NewService(createDomainRepo)),
		UpdateCompanyDomain: update_company_domain.NewHandler(update_company_domain.NewService(updateDomainRepo)),
		resolver:            NewResolver(db),
	}
}

func (c *Container) Resolver() middleware.CompanyResolver {
	return c.resolver
}

func (c *Container) RegisterRoutes(group *gin.RouterGroup, requirePermission func(string) gin.HandlerFunc, requireRole func(string) gin.HandlerFunc) {
	Register(group, c, requirePermission, requireRole)
}
