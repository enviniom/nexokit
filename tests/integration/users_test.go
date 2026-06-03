package integration_test

import (
	"bytes"
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

func TestUsersCRUDIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)
	db := helpers.NewSQLiteDB(t, &iamcore.IAMRole{}, &companies.Company{}, &iamcore.IAMUser{})
	company := helpers.SeedCompany(t, db, "acme")
	adminRole := helpers.SeedRole(t, db, iamcore.AdminRoleSlug)
	_ = helpers.SeedRole(t, db, iamcore.RootRoleSlug)

	actor := helpers.SeedUserWithRole(t, db, "actor-admin", adminRole, &company)
	container := iamusers.NewContainer(db, cache.NewNoop())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		authctx.SetGin(c, &authctx.User{PublicID: actor.PublicID, RoleSlug: iamcore.AdminRoleSlug, CompanyID: &company.ID, IsActive: true, Permissions: []string{"*"}})
		tenant.SetGin(c, tenant.NewScoped(company.ID, company.Slug))
		c.Next()
	})
	iamusers.Register(r.Group(""), container, middleware.RequirePermission)

	var created iamcore.UserResponse
	t.Run("create user", func(t *testing.T) {
		payload := map[string]any{"name": "John Tester", "email": "john@example.com", "password": "secret123", "role_id": adminRole.ID}
		w := requestJSON(r, http.MethodPost, "/users", payload)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp response.APIResponse[iamcore.UserResponse]
		mustDecode(t, w.Body.Bytes(), &resp)
		created = resp.Data
		if created.PublicID == "" {
			t.Fatalf("expected created user public id")
		}
	})

	t.Run("list includes created user in tenant scope", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp response.PaginatedResponse[[]iamcore.UserResponse]
		mustDecode(t, w.Body.Bytes(), &resp)
		if len(resp.Data) == 0 {
			t.Fatalf("expected non-empty users list")
		}
	})

	t.Run("update user", func(t *testing.T) {
		payload := map[string]any{"name": "John Updated", "email": "john.updated@example.com", "company_id": company.ID}
		w := requestJSON(r, http.MethodPut, "/users/"+created.PublicID, payload)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("delete user", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/users/"+created.PublicID, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}
		if w.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %q", w.Body.String())
		}

		fetch := httptest.NewRecorder()
		fetchReq := httptest.NewRequest(http.MethodGet, "/users/"+created.PublicID, bytes.NewReader(nil))
		r.ServeHTTP(fetch, fetchReq)
		if fetch.Code != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d: %s", fetch.Code, fetch.Body.String())
		}
	})

}
