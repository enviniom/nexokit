package view_role_permission_catalog

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryGetRoleByPublicID(t *testing.T) {
	db := mustOpenDB(t)
	company := seedCompany(t, db)
	seedRoleWithPermissions(t, db, company)

	repo := NewRepository(db)

	t.Run("returns tenant scoped role with permissions", func(t *testing.T) {
		role, err := repo.GetRoleByPublicID(tenant.NewScoped(company.ID, company.Slug), "role-admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if role.PublicID != "role-admin" {
			t.Fatalf("expected role-admin, got %s", role.PublicID)
		}
		if len(role.Permissions) != 2 {
			t.Fatalf("expected 2 permissions, got %d", len(role.Permissions))
		}
	})

	t.Run("maps missing role to domain not found", func(t *testing.T) {
		_, err := repo.GetRoleByPublicID(tenant.NewScoped(company.ID, company.Slug), "missing")
		if !errors.Is(err, core.ErrNotFound) {
			t.Fatalf("expected core.ErrNotFound, got %v", err)
		}
	})
}

func TestRepositoryListPermissionCatalog(t *testing.T) {
	db := mustOpenDB(t)
	seedPermissions(t, db)

	repo := NewRepository(db)
	items, err := repo.ListPermissionCatalog()
	if err != nil {
		t.Fatalf("list catalog: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 permissions, got %d", len(items))
	}

	if items[0].Slug != "roles.read" || items[1].Slug != "users.read" || items[2].Slug != "users.write" {
		t.Fatalf("unexpected order: [%s %s %s]", items[0].Slug, items[1].Slug, items[2].Slug)
	}
}

func mustOpenDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMRole{}, &core.IAMPermission{}, &core.IAMRolePermission{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func seedCompany(t *testing.T, db *gorm.DB) core.IAMCompany {
	t.Helper()
	company := core.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: "comp-a"}, Name: "Acme", Slug: "acme"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	return company
}

func seedPermissions(t *testing.T, db *gorm.DB) {
	t.Helper()
	permissions := []core.IAMPermission{
		{BaseModel: shared.BaseModel{PublicID: "perm-users-write"}, Slug: "users.write", Module: "users", Action: "write", Name: "Write Users", DisplayOrder: 2},
		{BaseModel: shared.BaseModel{PublicID: "perm-roles-read"}, Slug: "roles.read", Module: "roles", Action: "read", Name: "Read Roles", DisplayOrder: 1},
		{BaseModel: shared.BaseModel{PublicID: "perm-users-read"}, Slug: "users.read", Module: "users", Action: "read", Name: "Read Users", DisplayOrder: 1},
	}
	for i := range permissions {
		if err := db.Create(&permissions[i]).Error; err != nil {
			t.Fatalf("seed permission %s: %v", permissions[i].Slug, err)
		}
	}
}

func seedRoleWithPermissions(t *testing.T, db *gorm.DB, company core.IAMCompany) {
	t.Helper()
	seedPermissions(t, db)

	var permissions []core.IAMPermission
	if err := db.Order("slug ASC").Find(&permissions).Error; err != nil {
		t.Fatalf("load permissions: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-admin"}, Name: "Admin", Slug: "admin", CompanyID: &company.ID}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	if err := db.Model(&role).Association("Permissions").Append(&permissions[0], &permissions[1]); err != nil {
		t.Fatalf("assign permissions: %v", err)
	}
}
