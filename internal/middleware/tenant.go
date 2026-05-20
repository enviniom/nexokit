package middleware

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

const (
	companyHeader = "X-Company-ID"
	tenantHeader  = "X-Tenant"
	cacheTTL      = 5 * time.Minute
)

// ErrTenantNotFound marks an unresolved company lookup.
var ErrTenantNotFound = errors.New("tenant not found")

const msgCompanyScopeRequired = "X-Company-ID header is required for this tenant-scoped request"

// CompanyResolver resolves external company identifiers to internal tenant refs.
type CompanyResolver interface {
	FindByPublicIDOrSlug(value string) (tenant.CompanyRef, error)
	FindByHost(host string) (tenant.CompanyRef, error)
}

type cachedTenant struct {
	company tenant.CompanyRef
	expires time.Time
}

// RequireTenantScope resolves tenant context for tenant-owned private routes.
// Root users must provide X-Company-ID with a resolvable company public ID or slug.
func RequireTenantScope(resolver CompanyResolver) gin.HandlerFunc {
	return privateTenant(resolver, false)
}

// AllowRootGlobalScope resolves tenant context for platform/global private routes.
// Root users may omit X-Company-ID to operate with global scope.
func AllowRootGlobalScope(resolver CompanyResolver) gin.HandlerFunc {
	return privateTenant(resolver, true)
}

func privateTenant(resolver CompanyResolver, allowRootGlobal bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authctx.FromGin(c)
		if !ok || user == nil {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}

		if user.IsRoot {
			requestedCompany := strings.TrimSpace(c.GetHeader(companyHeader))
			if requestedCompany == "" {
				if !allowRootGlobal {
					response.BadRequest(c, msgCompanyScopeRequired)
					c.Abort()
					return
				}
				tenant.SetGin(c, tenant.NewRoot())
				c.Next()
				return
			}

			company, err := resolveByIDOrSlug(resolver, requestedCompany)
			if err != nil {
				response.BadRequest(c, messages.MsgBadRequest)
				c.Abort()
				return
			}
			tenant.SetGin(c, tenant.NewScoped(company.ID, company.Slug))
			c.Next()
			return
		}

		if !user.HasCompany() {
			response.Forbidden(c, messages.MsgForbidden)
			c.Abort()
			return
		}

		tenant.SetGin(c, tenant.NewScoped(*user.CompanyID, user.CompanySlug))
		c.Next()
	}
}

// PublicTenant resolves tenant context for unauthenticated public routes.
func PublicTenant(resolver CompanyResolver, appEnv string) gin.HandlerFunc {
	cache := map[string]cachedTenant{}

	return func(c *gin.Context) {
		company, ok := resolvePublicCompany(c, resolver, appEnv, cache)
		if !ok {
			response.NotFound(c, messages.MsgNotFound)
			c.Abort()
			return
		}

		tenant.SetGin(c, tenant.NewScoped(company.ID, company.Slug))
		c.Next()
	}
}

func resolvePublicCompany(c *gin.Context, resolver CompanyResolver, appEnv string, cache map[string]cachedTenant) (tenant.CompanyRef, bool) {
	host := normalizeHost(c.Request.Host)
	if host != "" {
		if company, ok := cachedLookup(cache, "host:"+host); ok {
			return company, true
		}
		if company, err := resolver.FindByHost(host); err == nil {
			cacheStore(cache, "host:"+host, company)
			return company, true
		}

		if slug := firstSubdomain(host); slug != "" {
			if company, ok := cachedLookup(cache, "slug:"+slug); ok {
				return company, true
			}
			if company, err := resolveByIDOrSlug(resolver, slug); err == nil {
				cacheStore(cache, "slug:"+slug, company)
				return company, true
			}
		}
	}

	if appEnv == "development" {
		requestedTenant := strings.TrimSpace(c.GetHeader(tenantHeader))
		if requestedTenant != "" {
			if company, ok := cachedLookup(cache, "slug:"+requestedTenant); ok {
				return company, true
			}
			if company, err := resolveByIDOrSlug(resolver, requestedTenant); err == nil {
				cacheStore(cache, "slug:"+requestedTenant, company)
				return company, true
			}
		}
	}

	return tenant.CompanyRef{}, false
}

func resolveByIDOrSlug(resolver CompanyResolver, value string) (tenant.CompanyRef, error) {
	if resolver == nil {
		return tenant.CompanyRef{}, ErrTenantNotFound
	}
	company, err := resolver.FindByPublicIDOrSlug(value)
	if err != nil {
		return tenant.CompanyRef{}, err
	}
	if company.ID == 0 {
		return tenant.CompanyRef{}, ErrTenantNotFound
	}
	return company, nil
}

func normalizeHost(host string) string {
	if host == "" {
		return ""
	}
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}
	return strings.ToLower(strings.TrimSpace(host))
}

func firstSubdomain(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) < 3 || parts[0] == "" || parts[0] == "www" {
		return ""
	}
	return parts[0]
}

func cachedLookup(cache map[string]cachedTenant, key string) (tenant.CompanyRef, bool) {
	entry, ok := cache[key]
	if !ok || time.Now().After(entry.expires) {
		return tenant.CompanyRef{}, false
	}
	return entry.company, true
}

func cacheStore(cache map[string]cachedTenant, key string, company tenant.CompanyRef) {
	cache[key] = cachedTenant{company: company, expires: time.Now().Add(cacheTTL)}
}
