package tenant

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const contextKey = "tenant_context"

// TenantContext carries the resolved company scope for a request.
//
// Values are intended to be created by tenant middleware via NewRoot or
// NewScoped and then passed by value to handlers and repositories.
type TenantContext struct {
	CompanyID   uint
	CompanySlug string
	IsRootScope bool
}

// CompanyRef is the minimum company data needed to build a tenant context.
type CompanyRef struct {
	ID   uint
	Slug string
}

// HostResolution describes the tenant and domain metadata resolved from a public host.
type HostResolution struct {
	Company           CompanyRef
	MatchedDomain     string
	DomainKind        string
	RedirectToPrimary bool
	PrimaryDomain     *string
}

// NewRoot returns a tenant context with global root scope.
func NewRoot() TenantContext {
	return TenantContext{IsRootScope: true}
}

// NewScoped returns a tenant context scoped to one company.
func NewScoped(companyID uint, companySlug string) TenantContext {
	return TenantContext{CompanyID: companyID, CompanySlug: companySlug}
}

// SetGin stores the tenant context in the Gin context.
func SetGin(c *gin.Context, tenant TenantContext) {
	c.Set(contextKey, tenant)
}

// FromGin retrieves the tenant context from the Gin context.
func FromGin(c *gin.Context) (TenantContext, bool) {
	value, exists := c.Get(contextKey)
	if !exists {
		return TenantContext{}, false
	}
	tenant, ok := value.(TenantContext)
	return tenant, ok
}

// WithCompany adds a company_id filter to a GORM query.
func WithCompany(db *gorm.DB, companyID uint) *gorm.DB {
	return db.Where("company_id = ?", companyID)
}

// ApplyTenantScope applies tenant filtering unless the request is root-global.
func ApplyTenantScope(db *gorm.DB, tenant TenantContext) *gorm.DB {
	if tenant.IsRootScope {
		return db
	}
	return WithCompany(db, tenant.CompanyID)
}
