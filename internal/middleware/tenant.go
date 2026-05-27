package middleware

import (
	"errors"
	"net"
	"net/http"
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

// HostResolution describes the tenant and domain metadata resolved from a public host.
type HostResolution = tenant.HostResolution

// CompanyResolver resolves external company identifiers to internal tenant refs.
type CompanyResolver interface {
	FindByPublicIDOrSlug(value string) (tenant.CompanyRef, error)
	ResolveHost(host string) (tenant.HostResolution, error)
}

type cachedTenant struct {
	company tenant.CompanyRef
	expires time.Time
}

type cachedHostResolution struct {
	resolution tenant.HostResolution
	expires    time.Time
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
	hostCache := map[string]cachedHostResolution{}
	tenantCache := map[string]cachedTenant{}

	return func(c *gin.Context) {
		resolution, ok := resolvePublicHost(c, resolver, hostCache)
		if ok {
			if redirectToPrimary(c, resolution) {
				return
			}
			tenant.SetGin(c, tenant.NewScoped(resolution.Company.ID, resolution.Company.Slug))
			c.Next()
			return
		}

		if appEnv == "development" {
			requestedTenant := strings.TrimSpace(c.GetHeader(tenantHeader))
			if requestedTenant != "" {
				if company, ok := cachedLookup(tenantCache, "slug:"+requestedTenant); ok {
					tenant.SetGin(c, tenant.NewScoped(company.ID, company.Slug))
					c.Next()
					return
				}
				if company, err := resolveByIDOrSlug(resolver, requestedTenant); err == nil {
					cacheStore(tenantCache, "slug:"+requestedTenant, company)
					tenant.SetGin(c, tenant.NewScoped(company.ID, company.Slug))
					c.Next()
					return
				}
			}
		}

		response.NotFound(c, messages.MsgNotFound)
		c.Abort()
	}
}

func resolvePublicHost(c *gin.Context, resolver CompanyResolver, cache map[string]cachedHostResolution) (tenant.HostResolution, bool) {
	host := normalizeHost(c.Request.Host)
	if host == "" || resolver == nil {
		return tenant.HostResolution{}, false
	}
	if resolution, ok := cachedHostLookup(cache, "host:"+host); ok {
		return resolution, true
	}
	resolution, err := resolver.ResolveHost(host)
	if err != nil || resolution.Company.ID == 0 {
		return tenant.HostResolution{}, false
	}
	cacheHostStore(cache, "host:"+host, resolution)
	return resolution, true
}

func redirectToPrimary(c *gin.Context, resolution tenant.HostResolution) bool {
	if !resolution.RedirectToPrimary || resolution.PrimaryDomain == nil {
		return false
	}
	primary := normalizeHost(*resolution.PrimaryDomain)
	matched := normalizeHost(resolution.MatchedDomain)
	if primary == "" || primary == matched {
		return false
	}
	location := requestScheme(c) + "://" + primary + c.Request.URL.RequestURI()
	c.Redirect(http.StatusPermanentRedirect, location)
	c.Abort()
	return true
}

func requestScheme(c *gin.Context) string {
	if proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		return strings.ToLower(strings.Split(proto, ",")[0])
	}
	if c.Request.TLS != nil {
		return "https"
	}
	if c.Request.URL != nil && c.Request.URL.Scheme != "" {
		return c.Request.URL.Scheme
	}
	return "http"
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

func cachedHostLookup(cache map[string]cachedHostResolution, key string) (tenant.HostResolution, bool) {
	entry, ok := cache[key]
	if !ok || time.Now().After(entry.expires) {
		return tenant.HostResolution{}, false
	}
	return entry.resolution, true
}

func cacheHostStore(cache map[string]cachedHostResolution, key string, resolution tenant.HostResolution) {
	cache[key] = cachedHostResolution{resolution: resolution, expires: time.Now().Add(cacheTTL)}
}
