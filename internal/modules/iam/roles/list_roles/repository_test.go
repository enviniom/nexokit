package list_roles

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryListAndCount(t *testing.T) {
	db := mustOpenDB(t)
	companyA, companyB := seedCompanies(t, db)
	seedRoles(t, db, companyA, companyB)

	repo := NewRepository(db)
	items, err := repo.List(tenant.NewScoped(companyA.ID, companyA.Slug), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 scoped roles, got %d", len(items))
	}
	var manager core.RoleResponse
	foundManager := false
	for _, item := range items {
		if item.CompanyID == nil || *item.CompanyID != companyA.PublicID {
			t.Fatalf("expected company public id %s in response", companyA.PublicID)
		}
		if item.Slug == "manager" {
			manager = item
			foundManager = true
		}
	}
	if !foundManager {
		t.Fatalf("expected manager role in result")
	}
	if len(manager.Permissions) != 2 || manager.Permissions[0] != "roles.read" || manager.Permissions[1] != "roles.write" {
		t.Fatalf("expected sorted permissions [roles.read roles.write], got %#v", manager.Permissions)
	}

	total, err := repo.Count(tenant.NewScoped(companyA.ID, companyA.Slug))
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
}

func TestGormRepositoryListSearchFilters(t *testing.T) {
	db := mustOpenDB(t)
	companyA, companyB := seedCompanies(t, db)
	seedRoles(t, db, companyA, companyB)

	repo := NewRepository(db)
	items, err := repo.List(tenant.NewRoot(), query.ListParams{
		Pagination: query.PaginationParams{Page: 1, PerPage: 10},
		Search:     query.SearchParams{Query: "auditor"},
	})
	if err != nil {
		t.Fatalf("list with search: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 searched role, got %d", len(items))
	}
	if items[0].Slug != "auditor" {
		t.Fatalf("expected slug auditor, got %s", items[0].Slug)
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
	read := core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-read"}, Slug: "roles.read", Name: "Read roles", Module: "roles", Action: "read"}
	write := core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-write"}, Slug: "roles.write", Name: "Write roles", Module: "roles", Action: "write"}
	if err := db.Create(&read).Error; err != nil {
		t.Fatalf("seed permission read: %v", err)
	}
	if err := db.Create(&write).Error; err != nil {
		t.Fatalf("seed permission write: %v", err)
	}

	manager := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-manager"}, Name: "Manager", Slug: "manager", CompanyID: &companyA.ID}
	auditor := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-auditor"}, Name: "Auditor", Slug: "auditor", CompanyID: &companyA.ID}
	viewer := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-viewer"}, Name: "Viewer", Slug: "viewer", CompanyID: &companyB.ID}
	if err := db.Create(&manager).Error; err != nil {
		t.Fatalf("seed role manager: %v", err)
	}
	if err := db.Model(&manager).Association("Permissions").Append(&write, &read); err != nil {
		t.Fatalf("seed role manager permissions: %v", err)
	}
	if err := db.Create(&auditor).Error; err != nil {
		t.Fatalf("seed role auditor: %v", err)
	}
	if err := db.Create(&viewer).Error; err != nil {
		t.Fatalf("seed role viewer: %v", err)
	}
}
