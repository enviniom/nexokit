package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/password"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUsersIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newUsersIsolationDB(t)
	rootRole := roles.Role{BaseModel: shared.BaseModel{ID: 1, PublicID: "role-root"}, Name: "root", Slug: roles.RootRoleSlug, IsSystem: true}
	adminRole := roles.Role{BaseModel: shared.BaseModel{ID: 2, PublicID: "role-admin"}, Name: "admin", Slug: roles.AdminRoleSlug, IsSystem: true}
	if err := db.Create(&rootRole).Error; err != nil {
		t.Fatalf("seed root role: %v", err)
	}
	if err := db.Create(&adminRole).Error; err != nil {
		t.Fatalf("seed admin role: %v", err)
	}
	companyOne := companies.Company{BaseModel: shared.BaseModel{ID: 1, PublicID: "company-acme"}, Name: "Acme", Slug: "acme", Status: companies.CompanyStatusActive}
	companyTwo := companies.Company{BaseModel: shared.BaseModel{ID: 2, PublicID: "company-globex"}, Name: "Globex", Slug: "globex", Status: companies.CompanyStatusActive}
	if err := db.Create(&companyOne).Error; err != nil {
		t.Fatalf("seed company one: %v", err)
	}
	if err := db.Create(&companyTwo).Error; err != nil {
		t.Fatalf("seed company two: %v", err)
	}
	seedUser(t, db, "root", "root@example.com", rootRole.ID, nil)
	seedUser(t, db, "admin-acme", "admin-acme@example.com", adminRole.ID, &companyOne.ID)
	seedUser(t, db, "user-acme", "user-acme@example.com", adminRole.ID, &companyOne.ID)
	seedUser(t, db, "user-globex", "user-globex@example.com", adminRole.ID, &companyTwo.ID)

	router := usersIsolationRouter(db)

	t.Run("admin list is isolated to own company", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		withActor(req, "admin", companyOne.ID)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeUsers(t, w)
		if len(got.Data) != 2 {
			t.Fatalf("expected 2 acme users, got %#v", got.Data)
		}
		for _, user := range got.Data {
			if user.CompanyID == nil || *user.CompanyID != companyOne.ID {
				t.Fatalf("expected only acme users, got %#v", got.Data)
			}
		}
	})

	t.Run("admin cross-tenant update returns not found", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"Globex Updated","email":"globex-updated@example.com","role_id":2,"company_id":2}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/users/user-globex", body)
		req.Header.Set("Content-Type", "application/json")
		withActor(req, "admin", companyOne.ID)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("root global list sees all users", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		withActor(req, "root", 0)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeUsers(t, w)
		if len(got.Data) != 4 {
			t.Fatalf("expected all 4 users for root, got %#v", got.Data)
		}
	})

	t.Run("root scoped list sees selected company", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		withActor(req, "root-scoped", companyTwo.ID)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeUsers(t, w)
		if len(got.Data) != 1 || got.Data[0].PublicID != "user-globex" {
			t.Fatalf("expected only scoped globex user, got %#v", got.Data)
		}
	})
}

func newUsersIsolationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&roles.Role{}, &companies.Company{}, &users.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func seedUser(t *testing.T, db *gorm.DB, publicID, email string, roleID uint, companyID *uint) {
	t.Helper()
	user := users.User{BaseModel: shared.BaseModel{PublicID: publicID}, Name: publicID, Email: email, PasswordHash: "hash", RoleID: roleID, CompanyID: companyID, IsActive: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user %s: %v", publicID, err)
	}
}

func usersIsolationRouter(db *gorm.DB) *gin.Engine {
	repo := users.NewRepository(db)
	rolesRepo := roles.NewRepository(db)
	svc := users.NewService(repo, password.Manager{}, roleResolverAdapter{repo: rolesRepo})
	handler := users.NewHandler(svc, authctx.PublicIDFromGin)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		actor := c.GetHeader("X-Test-Actor")
		switch actor {
		case "root":
			authctx.SetGin(c, &authctx.User{PublicID: "root", RoleSlug: roles.RootRoleSlug, IsRoot: true, IsActive: true, Permissions: []string{"*"}})
			tenant.SetGin(c, tenant.NewRoot())
		case "root-scoped":
			companyID := uint(2)
			authctx.SetGin(c, &authctx.User{PublicID: "root", RoleSlug: roles.RootRoleSlug, IsRoot: true, IsActive: true, Permissions: []string{"*"}})
			tenant.SetGin(c, tenant.NewScoped(companyID, "globex"))
		default:
			companyID := uint(1)
			authctx.SetGin(c, &authctx.User{PublicID: "admin-acme", RoleSlug: roles.AdminRoleSlug, CompanyID: &companyID, IsActive: true, Permissions: []string{"*"}})
			tenant.SetGin(c, tenant.NewScoped(companyID, "acme"))
		}
		c.Next()
	})
	users.Register(r.Group(""), handler, func(string) gin.HandlerFunc { return middleware.RequirePermission("*") })
	return r
}

func withActor(req *http.Request, actor string, companyID uint) {
	req.Header.Set("X-Test-Actor", actor)
	_ = companyID
}

func decodeUsers(t *testing.T, w *httptest.ResponseRecorder) response.APIResponse[[]users.UserResponse] {
	t.Helper()
	var resp response.APIResponse[[]users.UserResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}
