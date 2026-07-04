package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetRoleByPublicID(t *testing.T) {
	tests := []struct {
		name         string
		publicID     string
		seed         func(t *testing.T, db *gorm.DB)
		assertErr    func(t *testing.T, err error)
		assertResult func(t *testing.T, role *core.IAMRole)
	}{
		{
			name:     "returns role when found",
			publicID: "role-1",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
				role := core.IAMRole{
					BaseModel: shared.BaseModel{PublicID: "role-1"},
					Name:      "Admin",
					Slug:      "admin",
				}
				if err := db.Create(&role).Error; err != nil {
					t.Fatalf("seed role: %v", err)
				}
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			assertResult: func(t *testing.T, role *core.IAMRole) {
				t.Helper()
				if role.PublicID != "role-1" {
					t.Fatalf("expected public id %q, got %q", "role-1", role.PublicID)
				}
			},
		},
		{
			name:     "returns gorm not found when missing",
			publicID: "role-missing",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatalf("expected ErrRecordNotFound, got %v", err)
				}
			},
		},
		{
			name:     "preloads company association",
			publicID: "role-with-company",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMRole{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
				company := core.IAMCompany{
					BaseModelSimple: shared.BaseModelSimple{PublicID: "comp-1"},
					Name:            "Acme",
					Slug:            "acme",
				}
				if err := db.Create(&company).Error; err != nil {
					t.Fatalf("seed company: %v", err)
				}
				role := core.IAMRole{
					BaseModel: shared.BaseModel{PublicID: "role-with-company"},
					Name:      "Manager",
					Slug:      "manager",
					CompanyID: &company.ID,
				}
				if err := db.Create(&role).Error; err != nil {
					t.Fatalf("seed role: %v", err)
				}
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			assertResult: func(t *testing.T, role *core.IAMRole) {
				t.Helper()
				if role.Company.ID == 0 {
					t.Fatalf("expected company to be preloaded, got zero ID")
				}
				if role.Company.Slug != "acme" {
					t.Fatalf("expected company slug %q, got %q", "acme", role.Company.Slug)
				}
			},
		},
		{
			name:     "preloads permissions association",
			publicID: "role-with-permissions",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMPermission{}, &core.IAMRole{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
				permissions := []core.IAMPermission{
					{BaseModel: shared.BaseModel{PublicID: "perm-1"}, Slug: "users.read", Module: "users", Action: "read", Name: "Read", DisplayOrder: 1},
					{BaseModel: shared.BaseModel{PublicID: "perm-2"}, Slug: "users.write", Module: "users", Action: "write", Name: "Write", DisplayOrder: 2},
				}
				for i := range permissions {
					if err := db.Create(&permissions[i]).Error; err != nil {
						t.Fatalf("seed permission %s: %v", permissions[i].Slug, err)
					}
				}
				role := core.IAMRole{
					BaseModel:   shared.BaseModel{PublicID: "role-with-permissions"},
					Name:        "Operator",
					Slug:        "operator",
					Permissions: permissions,
				}
				if err := db.Create(&role).Error; err != nil {
					t.Fatalf("seed role: %v", err)
				}
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			assertResult: func(t *testing.T, role *core.IAMRole) {
				t.Helper()
				if len(role.Permissions) != 2 {
					t.Fatalf("expected 2 preloaded permissions, got %d", len(role.Permissions))
				}
				slugs := make(map[string]bool, len(role.Permissions))
				for _, p := range role.Permissions {
					slugs[p.Slug] = true
				}
				if !slugs["users.read"] || !slugs["users.write"] {
					t.Fatalf("expected both seeded permission slugs, got %+v", slugs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("open db: %v", err)
			}
			tt.seed(t, db)

			role, err := GetRoleByPublicID(db, tenant.NewRoot(), tt.publicID)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			if tt.assertResult != nil {
				tt.assertResult(t, role)
			}
		})
	}
}
