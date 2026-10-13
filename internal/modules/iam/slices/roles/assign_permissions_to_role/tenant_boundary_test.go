package assign_permissions_to_role

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/gin-gonic/gin"
)

func TestReplacementTenantBoundary(t *testing.T) {
	for _, name := range []string{"own", "foreign", "deleted", "missing context", "missing role", "root global", "root global null company", "root selected", "root selected foreign", "rollback"} {
		t.Run(name, func(t *testing.T) {
			db := mustOpenAssignDB(t)
			company, role, permissions := seedAssignData(t, db)
			tc := tenant.NewScoped(company.ID, company.Slug)
			wantDenied := false
			switch name {
			case "foreign", "root selected foreign":
				tc = tenant.NewScoped(company.ID+1, "other")
				wantDenied = true
			case "deleted":
				if err := db.Delete(&role).Error; err != nil {
					t.Fatal(err)
				}
				wantDenied = true
			case "missing context":
				tc = tenant.TenantContext{}
				wantDenied = true
			case "missing role":
				role.ID += 100
				wantDenied = true
			case "root global", "root global null company":
				tc = tenant.NewRoot()
				if name == "root global null company" {
					if err := db.Model(&role).Update("company_id", nil).Error; err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := db.Create(&iamrole.IAMRolePermission{RoleID: role.ID, PermissionID: permissions[0].ID}).Error; err != nil {
				t.Fatal(err)
			}
			ids := []uint{permissions[1].ID}
			if name == "rollback" {
				ids = append(ids, permissions[1].ID)
			}
			err := NewRepository(db).ReplacePermissions(tc, role.ID, ids)
			if wantDenied && !errors.Is(err, core.ErrNotFound) {
				t.Errorf("expected not found, got %v", err)
			}
			if !wantDenied && name != "rollback" && err != nil {
				t.Fatal(err)
			}
			if name == "rollback" && err == nil {
				t.Error("expected insert failure")
			}
			var links []iamrole.IAMRolePermission
			if err := db.Where("role_id = ?", role.ID).Find(&links).Error; err != nil {
				t.Fatal(err)
			}
			wantID := permissions[1].ID
			if wantDenied || name == "rollback" {
				wantID = permissions[0].ID
			}
			if len(links) != 1 || links[0].PermissionID != wantID {
				t.Errorf("association state changed unexpectedly: %+v", links)
			}
		})
	}
}

func TestMemberCacheTenantBoundary(t *testing.T) {
	for _, name := range []string{"own", "foreign", "missing context", "root global", "root selected"} {
		t.Run(name, func(t *testing.T) {
			db := mustOpenAssignDB(t)
			company, role, _ := seedAssignData(t, db)
			otherID := company.ID + 1
			foreign := iamuser.IAMUser{BaseModel: shared.BaseModel{PublicID: "foreign-user"}, CompanyID: &otherID, RoleID: role.ID, Name: "Foreign", Email: "foreign@test.dev", PasswordHash: "hash"}
			if err := db.Create(&foreign).Error; err != nil {
				t.Fatal(err)
			}
			tc := tenant.NewScoped(company.ID, company.Slug)
			want := 2
			switch name {
			case "foreign":
				tc = tenant.NewScoped(otherID, "other")
				want = 1
			case "missing context":
				tc = tenant.TenantContext{}
				want = 0
			case "root global":
				tc = tenant.NewRoot()
				want = 3
			}
			cache := &fakeCache{}
			err := NewRepository(db).InvalidateRoleMemberPermissionCache(tc, role.ID, cache)
			if name == "missing context" {
				if !errors.Is(err, core.ErrNotFound) {
					t.Fatalf("expected fail closed, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if len(cache.deleted) != want {
				t.Fatalf("expected %d scoped cache deletions, got %v", want, cache.deleted)
			}
			if name == "own" || name == "root selected" {
				for _, key := range cache.deleted {
					if key == "rbac:permissions:foreign-user" {
						t.Fatal("foreign member cache invalidated")
					}
				}
			}
		})
	}
}

// Use the same middleware order as the production private role routes.
func TestRoleAssignmentRouteActorBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	own := uint(7)
	for _, tt := range []struct {
		name   string
		actor  *authctx.User
		status int
	}{
		{"unauthenticated", nil, 401},
		{"spoofed display role", &authctx.User{Role: "admin", RoleSlug: "custom", CompanyID: &own}, 403},
		{"admin without company", &authctx.User{RoleSlug: roles.AdminRoleSlug}, 403},
		{"ordinary actor without grant", &authctx.User{RoleSlug: roles.UserRoleSlug, CompanyID: &own}, 403},
		{"tenant admin without grants", &authctx.User{RoleSlug: roles.AdminRoleSlug, CompanyID: &own}, 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.actor != nil {
					authctx.SetGin(c, tt.actor)
				}
			})
			group := r.Group("", middleware.AllowRootGlobalScope(nil))
			called := false
			group.PUT("/roles/:id/permissions", middleware.RequirePermission("roles.assign_permissions"), func(c *gin.Context) {
				tc, ok := tenant.FromGin(c)
				if !ok || tc.CompanyID != own || tc.IsRootScope {
					t.Fatal("actor-owned tenant scope missing")
				}
				called = true
				NewHandler(&fakeHandlerService{item: &iamrole.RolePermissionAssignmentResponse{}}).Handle(c)
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/roles/custom/permissions", bytes.NewBufferString(`{"permissions":[]}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Company-ID", "foreign")
			r.ServeHTTP(w, req)
			if w.Code != tt.status || called != (tt.status == 200) {
				t.Fatalf("status=%d called=%v, want %d", w.Code, called, tt.status)
			}
		})
	}
}
