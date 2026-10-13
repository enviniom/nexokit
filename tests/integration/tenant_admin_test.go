package integration_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/companies"
	companycore "github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/view_company"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/modules/iam/slices"
	iampermissions "github.com/enviniom/nexokit/internal/modules/iam/slices/permissions"
	"github.com/enviniom/nexokit/internal/modules/iam/slices/roles"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	core "github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantAdminProductionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&companycore.Company{}, &company_domain.CompanyDomain{}, &iamrole.IAMRole{}, &iamuser.IAMUser{}, &permission.IAMPermission{}); err != nil {
		t.Fatal(err)
	}
	own, foreign := uint(1), uint(2)
	for _, c := range []companycore.Company{{BaseModel: shared.BaseModel{ID: own, PublicID: "own-company"}, Name: "Own", Slug: "own"}, {BaseModel: shared.BaseModel{ID: foreign, PublicID: "foreign-company"}, Name: "Foreign", Slug: "foreign"}} {
		if err = db.Create(&c).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range []iamrole.IAMRole{
		{BaseModel: shared.BaseModel{ID: 1, PublicID: "admin-role"}, Name: "Tenant Administrator", Slug: core.AdminRoleSlug, CompanyID: &own, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 2, PublicID: "root-role"}, Name: "root", Slug: core.RootRoleSlug, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 3, PublicID: "own-custom"}, Name: "Custom", Slug: "custom", CompanyID: &own},
		{BaseModel: shared.BaseModel{ID: 4, PublicID: "foreign-custom"}, Name: "Foreign Custom", Slug: "foreign-custom", CompanyID: &foreign},
		{BaseModel: shared.BaseModel{ID: 5, PublicID: "own-system"}, Name: "System", Slug: "system", CompanyID: &own, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 6, PublicID: "own-user"}, Name: "user", Slug: core.UserRoleSlug, CompanyID: &own, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 7, PublicID: "foreign-user"}, Name: "user", Slug: core.UserRoleSlug, CompanyID: &foreign, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 8, PublicID: "foreign-system"}, Name: "Foreign System", Slug: "system", CompanyID: &foreign, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 9, PublicID: "ordinary-role"}, Name: "Ordinary", Slug: "ordinary", CompanyID: &own},
	} {
		if err = db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	seedUser(t, db, "tenant-admin", "tenant-admin@example.com", 1, &own)
	seedUser(t, db, "global-root", "global-root@example.com", 2, nil)
	p := permission.IAMPermission{BaseModel: shared.BaseModel{PublicID: "view-permission"}, Slug: "companies.view", Name: "View companies", Module: "companies", Action: "view", IsSystem: true}
	if err = db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	for _, roleID := range []uint{2, 4, 5, 6, 7, 8} {
		if err = db.Table("role_permissions").Create(map[string]any{"role_id": roleID, "permission_id": p.ID}).Error; err != nil {
			t.Fatal(err)
		}
	}
	seedUser(t, db, "ordinary-no-grants", "ordinary-no-grants@example.com", 9, &own)
	caps := slices.NewCapabilities(db, cache.NewNoop())
	actor, err := caps.AuthUserResolver.ResolveAuthUser("tenant-admin")
	if err != nil {
		t.Fatal(err)
	}
	if actor.RoleSlug != core.AdminRoleSlug || len(actor.Permissions) != 0 {
		t.Fatalf("expected DB-hydrated admin without grants: %#v", actor)
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		user, err := caps.AuthUserResolver.ResolveAuthUser(c.GetHeader("X-Test-User"))
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		authctx.SetGin(c, user)
	})
	globalGroup := r.Group("")
	globalGroup.Use(middleware.AllowRootGlobalScope(companies.NewResolver(db)))
	roles.Register(globalGroup, roles.NewContainer(db, cache.NewNoop()), middleware.RequirePermission)
	companies.Register(globalGroup, &companies.Container{ViewCompany: view_company.NewHandler(view_company.NewService(view_company.NewRepository(db)))}, middleware.RequirePermission, middleware.RequireRole)
	for _, tt := range []struct {
		name, actor, method, path, body string
		status                          int
	}{
		{"ordinary assignment denied", "ordinary-no-grants", "PUT", "/roles/own-user/permissions", `{"permissions":[]}`, 403},
		{"admin removes user system grants", "tenant-admin", "PUT", "/roles/own-user/permissions", `{"permissions":[]}`, 200},
		{"admin assigns user system grants", "tenant-admin", "PUT", "/roles/own-user/permissions", `{"permissions":["companies.view"]}`, 200},
		{"admin removes all user grants again", "tenant-admin", "PUT", "/roles/own-user/permissions", `{"permissions":[]}`, 200},
		{"foreign user hidden", "tenant-admin", "PUT", "/roles/foreign-user/permissions", `{"permissions":[]}`, 404},
		{"foreign system hidden", "tenant-admin", "PUT", "/roles/foreign-system/permissions", `{"permissions":[]}`, 404},
		{"root target immutable for root", "global-root", "PUT", "/roles/root-role/permissions", `{"permissions":[]}`, 403},
		{"admin target immutable for root", "global-root", "PUT", "/roles/admin-role/permissions", `{"permissions":["companies.view"]}`, 403},
		{"admin updates own custom", "tenant-admin", "PUT", "/roles/own-custom", `{"name":"Updated Custom","slug":"updated-custom"}`, 200},
		{"admin assigns own custom without grants", "tenant-admin", "PUT", "/roles/own-custom/permissions", `{"permissions":["companies.view"]}`, 200},
		{"foreign update hidden", "tenant-admin", "PUT", "/roles/foreign-custom", `{"name":"Changed","slug":"changed"}`, 404},
		{"foreign assignment hidden", "tenant-admin", "PUT", "/roles/foreign-custom/permissions", `{"permissions":[]}`, 404},
		{"admin update protected", "tenant-admin", "PUT", "/roles/admin-role", `{"name":"Changed","slug":"changed"}`, 403},
		{"admin assignment protected", "tenant-admin", "PUT", "/roles/admin-role/permissions", `{"permissions":[]}`, 403},
		{"system update protected", "tenant-admin", "PUT", "/roles/own-system", `{"name":"Changed","slug":"changed"}`, 403},
		{"system assignment editable", "tenant-admin", "PUT", "/roles/own-system/permissions", `{"permissions":["companies.view"]}`, 200},
		{"system assignment removable", "tenant-admin", "PUT", "/roles/own-system/permissions", `{"permissions":[]}`, 200},
		{"root role inaccessible", "tenant-admin", "PUT", "/roles/root-role/permissions", `{"permissions":[]}`, 404},
		{"root update inaccessible", "tenant-admin", "PUT", "/roles/root-role", `{"name":"Changed","slug":"changed"}`, 404},
		{"root delete inaccessible", "tenant-admin", "DELETE", "/roles/root-role", "", 404},
		{"own delete succeeds", "tenant-admin", "DELETE", "/roles/own-custom", "", 204},
		{"foreign delete hidden", "tenant-admin", "DELETE", "/roles/foreign-custom", "", 404},
		{"admin delete protected", "tenant-admin", "DELETE", "/roles/admin-role", "", 403},
		{"system delete protected", "tenant-admin", "DELETE", "/roles/own-system", "", 403},
		{"own company view", "tenant-admin", "GET", "/companies/own-company", "", 200},
		{"foreign company view hidden", "tenant-admin", "GET", "/companies/foreign-company", "", 404},
		{"root global company view", "global-root", "GET", "/companies/foreign-company", "", 200},
		{"root-only companies denied", "tenant-admin", "GET", "/companies", "", 403},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", tt.actor)
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
		})
	}
	t.Run("root selected header respects company scope", func(t *testing.T) {
		globalGroup.GET("/scope-probe", func(c *gin.Context) {
			tc, ok := tenant.FromGin(c)
			if !ok || tc.IsRootScope || tc.CompanyID != own {
				c.AbortWithStatus(500)
				return
			}
			c.Status(200)
		})
		for _, tt := range []struct {
			path, company string
			status        int
		}{
			{"/companies/foreign-company", "own-company", 404},
			{"/companies/foreign-company", "foreign-company", 200},
			{"/scope-probe", "own-company", 200},
		} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("X-Test-User", "global-root")
			req.Header.Set("X-Company-ID", tt.company)
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("%s: status=%d want=%d body=%s", tt.path, w.Code, tt.status, w.Body.String())
			}
		}
	})
	for _, tt := range []struct {
		roleID uint
		want   int64
	}{{1, 0}, {2, 1}, {4, 1}, {5, 0}, {6, 0}, {7, 1}, {8, 1}} {
		var count int64
		if err := db.Table("role_permissions").Where("role_id = ?", tt.roleID).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != tt.want {
			t.Errorf("role %d joins=%d want=%d", tt.roleID, count, tt.want)
		}
	}
	var role iamrole.IAMRole
	if err = db.Unscoped().Where("public_id = ?", "foreign-custom").First(&role).Error; err != nil {
		t.Fatal(err)
	}
	if role.Name != "Foreign Custom" || role.DeletedAt.Valid {
		t.Fatal("foreign role changed")
	}
}

func TestTenantPermissionCatalogMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newUsersIsolationDB(t)
	if err := db.AutoMigrate(&permission.IAMPermission{}); err != nil {
		t.Fatal(err)
	}
	own, other := uint(1), uint(2)
	for _, item := range []iamrole.IAMRole{
		{BaseModel: shared.BaseModel{ID: 1, PublicID: "admin-no-grants-role"}, Name: "Administrator", Slug: core.AdminRoleSlug, CompanyID: &own},
		{BaseModel: shared.BaseModel{ID: 2, PublicID: "admin-granted-role"}, Name: "Administrator", Slug: core.AdminRoleSlug, CompanyID: &other},
		{BaseModel: shared.BaseModel{ID: 3, PublicID: "ordinary-role"}, Name: "Ordinary", Slug: "ordinary", CompanyID: &own},
		{BaseModel: shared.BaseModel{ID: 4, PublicID: "root-catalog-role"}, Name: "root", Slug: core.RootRoleSlug, IsSystem: true},
		{BaseModel: shared.BaseModel{ID: 5, PublicID: "spoofed-catalog-role"}, Name: "root", Slug: "custom", CompanyID: &own},
	} {
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []permission.IAMPermission{
		{BaseModel: shared.BaseModel{PublicID: "global-custom"}, Name: "Original", Slug: "custom.view", Module: "custom", Action: "view"},
		{BaseModel: shared.BaseModel{PublicID: "global-system"}, Name: "System", Slug: "permissions.manage", Module: "permissions", Action: "manage", IsSystem: true},
	} {
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
	var system permission.IAMPermission
	if err := db.Where("public_id = ?", "global-system").First(&system).Error; err != nil {
		t.Fatal(err)
	}
	for _, roleID := range []uint{2, 3} {
		if err := db.Table("role_permissions").Create(map[string]any{"role_id": roleID, "permission_id": system.ID}).Error; err != nil {
			t.Fatal(err)
		}
	}
	seedUser(t, db, "admin-no-grants", "no-grants@example.com", 1, &own)
	seedUser(t, db, "admin-granted", "granted@example.com", 2, &other)
	seedUser(t, db, "ordinary-granted", "ordinary@example.com", 3, &own)
	seedUser(t, db, "catalog-root", "root@example.com", 4, nil)
	seedUser(t, db, "catalog-spoofed", "spoofed@example.com", 5, &own)
	caps := slices.NewCapabilities(db, cache.NewNoop())
	actor, err := caps.AuthUserResolver.ResolveAuthUser("admin-no-grants")
	if err != nil {
		t.Fatal(err)
	}
	if actor.RoleSlug != core.AdminRoleSlug || len(actor.Permissions) != 0 {
		t.Fatalf("expected trusted admin without grants: %#v", actor)
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		actor, err := caps.AuthUserResolver.ResolveAuthUser(c.GetHeader("X-Test-User"))
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		authctx.SetGin(c, actor)
	})
	group := r.Group("")
	group.Use(middleware.AllowRootGlobalScope(nil))
	iampermissions.Register(group, iampermissions.NewContainer(db), middleware.RequirePermission, middleware.RequireRole)
	for _, tt := range []struct {
		name, actor, method, path, wantName string
		status                              int
	}{
		{"unauthenticated cannot mutate global row", "", "PUT", "/permissions/global-custom", "Original", 401},
		{"root display name cannot mutate global row", "catalog-spoofed", "PUT", "/permissions/global-custom", "Original", 403},
		{"no-grant admin cannot mutate global row", "admin-no-grants", "PUT", "/permissions/global-custom", "Original", 403},
		{"granted admin cannot mutate global row", "admin-granted", "PUT", "/permissions/global-custom", "Original", 403},
		{"ordinary grant cannot mutate global row", "ordinary-granted", "PUT", "/permissions/global-custom", "Original", 403},
		{"admin can list catalog", "admin-no-grants", "GET", "/permissions", "Original", 200},
		{"admin can view catalog", "admin-no-grants", "GET", "/permissions/global-custom", "Original", 200},
		{"root can mutate non-system catalog", "catalog-root", "PUT", "/permissions/global-custom", "Changed", 200},
		{"root cannot mutate system catalog", "catalog-root", "PUT", "/permissions/global-system", "Changed", 403},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(`{"name":"Changed","description":"changed","display_order":9}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", tt.actor)
			r.ServeHTTP(w, req)
			var row permission.IAMPermission
			if err := db.Where("public_id = ?", "global-custom").First(&row).Error; err != nil {
				t.Fatal(err)
			}
			if row.Name != tt.wantName {
				t.Errorf("persisted name = %q, want %q", row.Name, tt.wantName)
			}
			if tt.wantName == "Original" && (row.Description != "" || row.DisplayOrder != 0) {
				t.Errorf("denied mutation changed global row: %#v", row)
			}
			if w.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
		})
	}
	if err := db.Where("public_id = ?", "global-system").First(&system).Error; err != nil {
		t.Fatal(err)
	}
	if system.Name != "System" || system.Description != "" || system.DisplayOrder != 0 || !system.IsSystem {
		t.Fatal("system catalog row changed")
	}
}
