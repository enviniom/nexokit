package list_selectable_roles

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryListSelectableRoles(t *testing.T) {
	db := mustOpenDB(t)
	companyA, companyB := seedCompanies(t, db)
	seedRoles(t, db, companyA, companyB)

	repo := NewRepository(db)

	t.Run("tenant scoped excludes root and other tenant roles", func(t *testing.T) {
		items, err := repo.List(tenant.NewScoped(companyA.ID, companyA.Slug))
		if err != nil {
			t.Fatalf("list scoped: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 roles, got %d", len(items))
		}
		if items[0].Name != "Admin" || items[1].Name != "Manager" {
			t.Fatalf("expected sorted names [Admin Manager], got [%s %s]", items[0].Name, items[1].Name)
		}
		if items[0].Meta["company_id"] != companyA.PublicID {
			t.Fatalf("expected company_id meta %s", companyA.PublicID)
		}
	})

	t.Run("root scope includes cross-tenant non-root roles", func(t *testing.T) {
		items, err := repo.List(tenant.NewRoot())
		if err != nil {
			t.Fatalf("list root: %v", err)
		}
		if len(items) != 4 {
			t.Fatalf("expected 4 non-root roles, got %d", len(items))
		}
		for _, item := range items {
			if item.Meta["slug"] == core.RootRoleSlug {
				t.Fatalf("root role must not be selectable")
			}
		}
	})
}

func mustOpenDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func seedCompanies(t *testing.T, db *gorm.DB) (core.IAMCompany, core.IAMCompany) {
	t.Helper()
	companyA := core.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: "comp-a"}, Name: "Acme", Slug: "acme"}
	companyB := core.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: "comp-b"}, Name: "Beta", Slug: "beta"}
	if err := db.Create(&companyA).Error; err != nil {
		t.Fatalf("seed companyA: %v", err)
	}
	if err := db.Create(&companyB).Error; err != nil {
		t.Fatalf("seed companyB: %v", err)
	}
	return companyA, companyB
}

func seedRoles(t *testing.T, db *gorm.DB, companyA, companyB core.IAMCompany) {
	t.Helper()
	roles := []core.IAMRole{
		{BaseModel: shared.BaseModel{PublicID: "role-root"}, Name: "Root", Slug: core.RootRoleSlug},
		{BaseModel: shared.BaseModel{PublicID: "role-admin-a"}, Name: "Admin", Slug: "admin", CompanyID: &companyA.ID, Company: companyA},
		{BaseModel: shared.BaseModel{PublicID: "role-manager-a"}, Name: "Manager", Slug: "manager", CompanyID: &companyA.ID, Company: companyA},
		{BaseModel: shared.BaseModel{PublicID: "role-admin-b"}, Name: "Admin B", Slug: "admin-b", CompanyID: &companyB.ID, Company: companyB},
		{BaseModel: shared.BaseModel{PublicID: "role-viewer-b"}, Name: "Viewer", Slug: "viewer", CompanyID: &companyB.ID, Company: companyB},
	}
	for i := range roles {
		if err := db.Create(&roles[i]).Error; err != nil {
			t.Fatalf("seed role %s: %v", roles[i].Slug, err)
		}
	}
}
