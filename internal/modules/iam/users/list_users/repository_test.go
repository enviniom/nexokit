package list_users

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
	role := seedRole(t, db)
	seedUsers(t, db, companyA, companyB, role)

	repo := NewRepository(db)
	tc := tenant.NewScoped(companyA.ID, companyA.Slug)
	items, err := repo.List(tc, query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 scoped users, got %d", len(items))
	}
	for _, item := range items {
		if item.CompanyID == nil || *item.CompanyID != companyA.ID {
			t.Fatalf("expected company_id %d in response, got %v", companyA.ID, item.CompanyID)
		}
		if item.RoleName != "Admin" {
			t.Fatalf("expected role name Admin, got %s", item.RoleName)
		}
	}

	total, err := repo.Count(tc, query.ListParams{})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
}

func TestGormRepositoryListSearchAndStatusFilter(t *testing.T) {
	db := mustOpenDB(t)
	companyA, companyB := seedCompanies(t, db)
	role := seedRole(t, db)
	seedUsers(t, db, companyA, companyB, role)

	repo := NewRepository(db)
	tc := tenant.NewRoot()

	t.Run("search by name", func(t *testing.T) {
		items, err := repo.List(tc, query.ListParams{
			Pagination: query.PaginationParams{Page: 1, PerPage: 10},
			Search:     query.SearchParams{Query: "Alice"},
		})
		if err != nil {
			t.Fatalf("list with search: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 searched user, got %d", len(items))
		}
		if items[0].Name != "Alice" {
			t.Fatalf("expected name Alice, got %s", items[0].Name)
		}
	})

	t.Run("filter active", func(t *testing.T) {
		items, err := repo.List(tc, query.ListParams{
			Pagination: query.PaginationParams{Page: 1, PerPage: 10},
			Filters:    query.FilterParams{Status: "active"},
		})
		if err != nil {
			t.Fatalf("list with active filter: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 active users, got %d", len(items))
		}
	})

	t.Run("filter inactive", func(t *testing.T) {
		items, err := repo.List(tc, query.ListParams{
			Pagination: query.PaginationParams{Page: 1, PerPage: 10},
			Filters:    query.FilterParams{Status: "inactive"},
		})
		if err != nil {
			t.Fatalf("list with inactive filter: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 inactive user, got %d", len(items))
		}
		if items[0].Name != "Charlie" {
			t.Fatalf("expected Charlie, got %s", items[0].Name)
		}
	})

	t.Run("count with filter", func(t *testing.T) {
		total, err := repo.Count(tc, query.ListParams{
			Filters: query.FilterParams{Status: "active"},
		})
		if err != nil {
			t.Fatalf("count with filter: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected filtered count 2, got %d", total)
		}
	})
}

func TestGormRepositoryListTenantIsolation(t *testing.T) {
	db := mustOpenDB(t)
	companyA, companyB := seedCompanies(t, db)
	role := seedRole(t, db)
	seedUsers(t, db, companyA, companyB, role)

	repo := NewRepository(db)
	tcB := tenant.NewScoped(companyB.ID, companyB.Slug)
	items, err := repo.List(tcB, query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 user for companyB, got %d", len(items))
	}
	if items[0].Name != "Charlie" {
		t.Fatalf("expected Charlie, got %s", items[0].Name)
	}
}

func TestToUserResponse(t *testing.T) {
	companyID := uint(42)
	user := &core.IAMUser{
		BaseModel: shared.BaseModel{PublicID: "user-abc"},
		Name:      "Test",
		Email:     "test@example.com",
		IsActive:  true,
		RoleID:    1,
		CompanyID: &companyID,
		Role:      core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin"},
	}
	resp := toUserResponse(user)
	if resp.PublicID != "user-abc" {
		t.Fatalf("expected public_id user-abc, got %s", resp.PublicID)
	}
	if resp.RoleName != "Admin" {
		t.Fatalf("expected role name Admin, got %s", resp.RoleName)
	}
	if resp.CompanyID == nil || *resp.CompanyID != 42 {
		t.Fatalf("expected company_id 42, got %v", resp.CompanyID)
	}
}

func mustOpenDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMRole{}, &core.IAMUser{}); err != nil {
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

func seedRole(t *testing.T, db *gorm.DB) core.IAMRole {
	t.Helper()
	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-admin"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

func seedUsers(t *testing.T, db *gorm.DB, companyA, companyB core.IAMCompany, role core.IAMRole) {
	t.Helper()
	alice := core.IAMUser{
		BaseModel: shared.BaseModel{PublicID: "user-alice"}, Name: "Alice", Email: "alice@example.com",
		PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyA.ID, IsActive: true,
	}
	bob := core.IAMUser{
		BaseModel: shared.BaseModel{PublicID: "user-bob"}, Name: "Bob", Email: "bob@example.com",
		PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyA.ID, IsActive: true,
	}
	charlie := core.IAMUser{
		BaseModel: shared.BaseModel{PublicID: "user-charlie"}, Name: "Charlie", Email: "charlie@example.com",
		PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyB.ID, IsActive: true,
	}
	for _, u := range []core.IAMUser{alice, bob, charlie} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("seed user %s: %v", u.Name, err)
		}
	}
	// GORM skips zero-value bools on Create when a default exists, so explicitly deactivate Charlie.
	if err := db.Model(&core.IAMUser{}).Where("public_id = ?", "user-charlie").Update("is_active", false).Error; err != nil {
		t.Fatalf("deactivate charlie: %v", err)
	}
}
