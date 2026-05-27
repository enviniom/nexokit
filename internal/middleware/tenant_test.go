package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeCompanyResolver struct {
	byIDOrSlug map[string]tenant.CompanyRef
	byHost     map[string]HostResolution
	err        error
}

func (f fakeCompanyResolver) FindByPublicIDOrSlug(value string) (tenant.CompanyRef, error) {
	if f.err != nil {
		return tenant.CompanyRef{}, f.err
	}
	company, ok := f.byIDOrSlug[value]
	if !ok {
		return tenant.CompanyRef{}, ErrTenantNotFound
	}
	return company, nil
}

func (f fakeCompanyResolver) ResolveHost(host string) (HostResolution, error) {
	if f.err != nil {
		return HostResolution{}, f.err
	}
	resolution, ok := f.byHost[host]
	if !ok {
		return HostResolution{}, ErrTenantNotFound
	}
	return resolution, nil
}

func TestAllowRootGlobalScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		user       *authctx.User
		header     string
		resolver   CompanyResolver
		wantStatus int
		wantTenant tenant.TenantContext
	}{
		{
			name:       "root without header gets global tenant scope",
			user:       &authctx.User{RoleSlug: "root", IsRoot: true, IsActive: true},
			resolver:   fakeCompanyResolver{},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewRoot(),
		},
		{
			name:   "root with company header gets scoped tenant",
			user:   &authctx.User{RoleSlug: "root", IsRoot: true, IsActive: true},
			header: "company_01",
			resolver: fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{
				"company_01": {ID: 7, Slug: "acme"},
			}},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(7, "acme"),
		},
		{
			name:       "admin gets company scope from authenticated user",
			user:       &authctx.User{RoleSlug: "admin", CompanyID: uintPtr(4), IsActive: true},
			resolver:   fakeCompanyResolver{},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(4, ""),
		},
		{
			name:       "admin company header is ignored",
			user:       &authctx.User{RoleSlug: "admin", CompanyID: uintPtr(4), IsActive: true},
			header:     "company_07",
			resolver:   fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{"company_07": {ID: 7, Slug: "globex"}}},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(4, ""),
		},
		{
			name:       "non-root without company is forbidden",
			user:       &authctx.User{RoleSlug: "admin", IsActive: true},
			resolver:   fakeCompanyResolver{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "root with unknown company header is bad request",
			user:       &authctx.User{RoleSlug: "root", IsRoot: true, IsActive: true},
			header:     "missing",
			resolver:   fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{}},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				authctx.SetGin(c, tt.user)
				c.Next()
			})
			r.Use(AllowRootGlobalScope(tt.resolver))
			r.GET("/private", tenantEchoHandler(t))

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tt.header != "" {
				req.Header.Set("X-Company-ID", tt.header)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				assertTenant(t, w, tt.wantTenant)
			}
		})
	}
}

func TestRequireTenantScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		user       *authctx.User
		header     string
		resolver   CompanyResolver
		wantStatus int
		wantTenant tenant.TenantContext
	}{
		{
			name:       "root without header is rejected",
			user:       &authctx.User{RoleSlug: "root", IsRoot: true, IsActive: true},
			resolver:   fakeCompanyResolver{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "root with company header gets scoped tenant",
			user:   &authctx.User{RoleSlug: "root", IsRoot: true, IsActive: true},
			header: "company_01",
			resolver: fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{
				"company_01": {ID: 7, Slug: "acme"},
			}},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(7, "acme"),
		},
		{
			name:       "admin gets company scope from authenticated user",
			user:       &authctx.User{RoleSlug: "admin", CompanyID: uintPtr(4), IsActive: true},
			resolver:   fakeCompanyResolver{},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(4, ""),
		},
		{
			name:       "admin company header is ignored",
			user:       &authctx.User{RoleSlug: "admin", CompanyID: uintPtr(4), IsActive: true},
			header:     "company_07",
			resolver:   fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{"company_07": {ID: 7, Slug: "globex"}}},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(4, ""),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				authctx.SetGin(c, tt.user)
				c.Next()
			})
			r.Use(RequireTenantScope(tt.resolver))
			r.GET("/private", tenantEchoHandler(t))

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tt.header != "" {
				req.Header.Set("X-Company-ID", tt.header)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				assertTenant(t, w, tt.wantTenant)
			}
		})
	}
}

func TestPublicTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		env        string
		host       string
		header     string
		resolver   CompanyResolver
		wantStatus int
		wantTenant tenant.TenantContext
	}{
		{
			name: "host resolves exact active company domain",
			env:  "production",
			host: "store.acme.com",
			resolver: fakeCompanyResolver{byHost: map[string]HostResolution{
				"store.acme.com": {Company: tenant.CompanyRef{ID: 11, Slug: "acme"}, MatchedDomain: "store.acme.com"},
			}},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(11, "acme"),
		},
		{
			name: "subdomain does not resolve company slug without explicit domain row",
			env:  "production",
			host: "acme.app.nexokit.com",
			resolver: fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{
				"acme": {ID: 12, Slug: "acme"},
			}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "development x-tenant resolves company slug",
			env:    "development",
			header: "devco",
			resolver: fakeCompanyResolver{byIDOrSlug: map[string]tenant.CompanyRef{
				"devco": {ID: 13, Slug: "devco"},
			}},
			wantStatus: http.StatusOK,
			wantTenant: tenant.NewScoped(13, "devco"),
		},
		{
			name:       "production ignores x-tenant and returns not found",
			env:        "production",
			header:     "devco",
			resolver:   fakeCompanyResolver{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "resolver errors return not found for public routes",
			env:        "production",
			host:       "unknown.example.com",
			resolver:   fakeCompanyResolver{err: errors.New("database unavailable")},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(PublicTenant(tt.resolver, tt.env))
			r.GET("/public", tenantEchoHandler(t))

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/public", nil)
			req.Host = tt.host
			if tt.header != "" {
				req.Header.Set("X-Tenant", tt.header)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				assertTenant(t, w, tt.wantTenant)
			}
		})
	}
}

func TestPublicTenantRedirectsAliasToPrimaryDomain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	primary := "dreammakers.com.co"
	r := gin.New()
	r.Use(PublicTenant(fakeCompanyResolver{byHost: map[string]HostResolution{
		"www.dreammakers.com.co": {
			Company:           tenant.CompanyRef{ID: 21, Slug: "dreammakers"},
			MatchedDomain:     "www.dreammakers.com.co",
			RedirectToPrimary: true,
			PrimaryDomain:     &primary,
		},
	}}, "production"))
	r.GET("/products", tenantEchoHandler(t))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://www.dreammakers.com.co/products?tag=sale", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusPermanentRedirect, w.Body.String())
	}
	if got := w.Header().Get("Location"); got != "https://dreammakers.com.co/products?tag=sale" {
		t.Fatalf("Location = %q, want %q", got, "https://dreammakers.com.co/products?tag=sale")
	}
}

func tenantEchoHandler(t *testing.T) gin.HandlerFunc {
	t.Helper()
	return func(c *gin.Context) {
		tc, ok := tenant.FromGin(c)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		response.Success(c, "ok", tc)
	}
}

func assertTenant(t *testing.T, w *httptest.ResponseRecorder, want tenant.TenantContext) {
	t.Helper()
	var got response.APIResponse[tenant.TenantContext]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal tenant response: %v; body=%s", err, w.Body.String())
	}
	if got.Data != want {
		t.Fatalf("tenant = %#v, want %#v", got.Data, want)
	}
}

func uintPtr(value uint) *uint {
	return &value
}
