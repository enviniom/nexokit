package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/companies"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	iamusers "github.com/enviniom/nexokit/internal/modules/iam/users"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/tests/helpers"
	"github.com/gin-gonic/gin"
)

func TestTenantIsolationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)
	db := helpers.NewSQLiteDB(t, &iamcore.IAMRole{}, &companies.Company{}, &iamcore.IAMUser{})
	rootRole := helpers.SeedRole(t, db, iamcore.RootRoleSlug)
	adminRole := helpers.SeedRole(t, db, iamcore.AdminRoleSlug)
	companyA := helpers.SeedCompany(t, db, "acme")
	companyB := helpers.SeedCompany(t, db, "globex")

	_ = helpers.SeedUserWithRole(t, db, "admin-a", adminRole, &companyA)
	_ = helpers.SeedUserWithRole(t, db, "user-a", adminRole, &companyA)
	userB := helpers.SeedUserWithRole(t, db, "user-b", adminRole, &companyB)
	root := helpers.SeedUserWithRole(t, db, "root-user", rootRole, nil)

	container := iamusers.NewContainer(db, cache.NewNoop())
	r := gin.New()
	r.Use(func(c *gin.Context) {
		actor := c.GetHeader("X-Actor")
		switch actor {
		case "root":
			authctx.SetGin(c, &authctx.User{PublicID: root.PublicID, RoleSlug: iamcore.RootRoleSlug, IsRoot: true, IsActive: true, Permissions: []string{"*"}})
			tenant.SetGin(c, tenant.NewRoot())
		case "root-scoped":
			authctx.SetGin(c, &authctx.User{PublicID: root.PublicID, RoleSlug: iamcore.RootRoleSlug, IsRoot: true, IsActive: true, Permissions: []string{"*"}})
			tenant.SetGin(c, tenant.NewScoped(companyB.ID, companyB.Slug))
		default:
			authctx.SetGin(c, &authctx.User{PublicID: "admin-a", RoleSlug: iamcore.AdminRoleSlug, CompanyID: &companyA.ID, IsActive: true, Permissions: []string{"*"}})
			tenant.SetGin(c, tenant.NewScoped(companyA.ID, companyA.Slug))
		}
		c.Next()
	})
	iamusers.Register(r.Group(""), container, middleware.RequirePermission)

	t.Run("admin cannot access other company user", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/"+userB.PublicID, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-tenant lookup, got %d", w.Code)
		}
	})

	t.Run("root global sees all users", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("X-Actor", "root")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp response.PaginatedResponse[[]iamcore.UserResponse]
		mustDecode(t, w.Body.Bytes(), &resp)
		if len(resp.Data) < 4 {
			t.Fatalf("expected root to see all users, got %d", len(resp.Data))
		}
	})

	t.Run("root scoped sees selected company only", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("X-Actor", "root-scoped")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp response.PaginatedResponse[[]iamcore.UserResponse]
		mustDecode(t, w.Body.Bytes(), &resp)
		for _, u := range resp.Data {
			if u.CompanyID == nil || *u.CompanyID != companyB.ID {
				t.Fatalf("expected only company B users, got %+v", resp.Data)
			}
		}
	})
}
