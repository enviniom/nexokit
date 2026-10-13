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

type fakePermissionResolver struct {
	permissions []string
	err         error
	calledWith  string
	calls       int
}

func (f *fakePermissionResolver) Resolve(publicID string) ([]string, error) {
	f.calledWith = publicID
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.permissions, nil
}

func TestAttachPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("loads permissions on cache miss and stores them in auth context", func(t *testing.T) {
		resolver := &fakePermissionResolver{permissions: []string{"users.list", "roles.view"}}
		r := gin.New()
		r.Use(func(c *gin.Context) {
			authctx.SetGin(c, &authctx.User{PublicID: "user1", Role: "admin", RoleSlug: "admin", IsActive: true})
			c.Next()
		})
		r.Use(AttachPermissions(resolver))
		r.GET("/protected", func(c *gin.Context) {
			user, ok := authctx.FromGin(c)
			if !ok || user == nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			response.Success(c, "ok", gin.H{"permissions": user.Permissions})
		})

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if resolver.calledWith != "user1" || resolver.calls != 1 {
			t.Fatalf("expected resolver to load user1 once, got %q x%d", resolver.calledWith, resolver.calls)
		}
		var resp response.APIResponse[struct {
			Permissions []string `json:"permissions"`
		}]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Data.Permissions) != 2 || resp.Data.Permissions[0] != "users.list" || resp.Data.Permissions[1] != "roles.view" {
			t.Fatalf("unexpected permissions: %#v", resp.Data.Permissions)
		}
	})

	t.Run("resolver failure degrades to empty permissions", func(t *testing.T) {
		resolver := &fakePermissionResolver{err: errors.New("cache down")}
		r := gin.New()
		r.Use(func(c *gin.Context) {
			authctx.SetGin(c, &authctx.User{PublicID: "user1", Role: "admin", RoleSlug: "admin", Permissions: []string{"stale.permission"}, IsActive: true})
			c.Next()
		})
		r.Use(AttachPermissions(resolver))
		r.GET("/protected", func(c *gin.Context) {
			user, _ := authctx.FromGin(c)
			response.Success(c, "ok", gin.H{"permissions": user.Permissions})
		})

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("expected graceful 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp response.APIResponse[struct {
			Permissions []string `json:"permissions"`
		}]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Data.Permissions) != 0 {
			t.Fatalf("expected empty permissions after resolver failure, got %#v", resp.Data.Permissions)
		}
	})
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		user   *authctx.User
		status int
	}{
		{name: "unauthenticated returns 401", status: http.StatusUnauthorized},
		{name: "missing permission returns 403", user: &authctx.User{PublicID: "user1", Role: "user", RoleSlug: "user", Permissions: []string{"users.view"}}, status: http.StatusForbidden},
		{name: "matching permission proceeds", user: &authctx.User{PublicID: "user1", Role: "admin", RoleSlug: "admin", Permissions: []string{"users.create"}}, status: http.StatusOK},
		{name: "root role bypasses permission check", user: &authctx.User{PublicID: "root1", Role: "root", RoleSlug: "root", IsRoot: true}, status: http.StatusOK},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			if tt.user != nil {
				r.Use(func(c *gin.Context) {
					authctx.SetGin(c, tt.user)
					c.Next()
				})
			}
			r.GET("/protected", RequirePermission("users.create"), func(c *gin.Context) { response.Success(c, "ok", gin.H{"allowed": true}) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

			if w.Code != tt.status {
				t.Fatalf("expected status %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
			var resp response.APIResponse[any]
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("expected standard envelope: %v", err)
			}
			if tt.status == http.StatusOK && !resp.Success {
				t.Fatalf("expected success envelope, got %#v", resp)
			}
			if tt.status != http.StatusOK && resp.Success {
				t.Fatalf("expected error envelope, got %#v", resp)
			}
		})
	}
}

func TestTenantAdminPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uint(7)
	cases := []struct {
		name   string
		user   *authctx.User
		scope  *tenant.TenantContext
		status int
	}{
		{"own tenant without grants", &authctx.User{RoleSlug: "admin", CompanyID: &id}, scopePtr(tenant.NewScoped(id, "own")), 200},
		{"missing scope", &authctx.User{RoleSlug: "admin", CompanyID: &id}, nil, 403},
		{"global scope", &authctx.User{RoleSlug: "admin", CompanyID: &id}, scopePtr(tenant.NewRoot()), 403},
		{"foreign scope", &authctx.User{RoleSlug: "admin", CompanyID: &id}, scopePtr(tenant.NewScoped(8, "other")), 403},
		{"zero scope", &authctx.User{RoleSlug: "admin", CompanyID: &id}, scopePtr(tenant.NewScoped(0, "")), 403},
		{"no company", &authctx.User{RoleSlug: "admin"}, scopePtr(tenant.NewScoped(id, "own")), 403},
		{"legacy spoof", &authctx.User{Role: "admin", CompanyID: &id}, scopePtr(tenant.NewScoped(id, "own")), 403},
		{"ordinary denied", &authctx.User{RoleSlug: "user", CompanyID: &id}, scopePtr(tenant.NewScoped(id, "own")), 403},
		{"ordinary granted", &authctx.User{RoleSlug: "user", CompanyID: &id, Permissions: []string{"users.create"}}, scopePtr(tenant.NewScoped(id, "own")), 200},
		{"root global", &authctx.User{IsRoot: true}, scopePtr(tenant.NewRoot()), 200},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				authctx.SetGin(c, tt.user)
				if tt.scope != nil {
					tenant.SetGin(c, *tt.scope)
				}
			})
			r.GET("/", RequirePermission("users.create"), func(c *gin.Context) { c.Status(200) })
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			if w.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
		})
	}
	t.Run("admin cannot use root-only route", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			authctx.SetGin(c, &authctx.User{RoleSlug: "admin", CompanyID: &id})
			tenant.SetGin(c, tenant.NewScoped(id, "own"))
		})
		r.GET("/", RequireRole("root"), func(c *gin.Context) { c.Status(200) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != 403 {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})
}
func TestRootRoleGuardRequiresTrustedRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name   string
		actor  *authctx.User
		status int
	}{
		{"missing actor", nil, 401},
		{"legacy root text", &authctx.User{Role: "root"}, 403},
		{"slug without trusted root", &authctx.User{RoleSlug: "root"}, 403},
		{"admin with catalog grant", &authctx.User{RoleSlug: "admin", Permissions: []string{"permissions.manage"}}, 403},
		{"trusted root", &authctx.User{RoleSlug: "root", IsRoot: true}, 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.actor != nil {
					authctx.SetGin(c, tt.actor)
				}
			})
			r.PUT("/permissions/:id", RequireRole("root"), func(c *gin.Context) { c.Status(200) })
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("PUT", "/permissions/custom", nil))
			if w.Code != tt.status {
				t.Fatalf("got %d, want %d", w.Code, tt.status)
			}
		})
	}
}

func scopePtr(tc tenant.TenantContext) *tenant.TenantContext { return &tc }

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		user   *authctx.User
		status int
	}{
		{name: "unauthenticated returns 401", status: http.StatusUnauthorized},
		{name: "role mismatch returns 403", user: &authctx.User{PublicID: "user1", Role: "user", RoleSlug: "user"}, status: http.StatusForbidden},
		{name: "role match proceeds", user: &authctx.User{PublicID: "admin1", Role: "admin", RoleSlug: "admin"}, status: http.StatusOK},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			if tt.user != nil {
				r.Use(func(c *gin.Context) {
					authctx.SetGin(c, tt.user)
					c.Next()
				})
			}
			r.GET("/protected", RequireRole("admin"), func(c *gin.Context) { response.Success(c, "ok", gin.H{"allowed": true}) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

			if w.Code != tt.status {
				t.Fatalf("expected status %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
		})
	}
}
