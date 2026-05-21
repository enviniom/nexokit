package companies

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_ListAppliesFiltersSearchSortingAndPagination(t *testing.T) {
	db := newCompanyRepositoryTestDB(t)
	from := mustCompanyRepositoryDate(t, "2025-01-01")
	to := mustCompanyRepositoryDate(t, "2025-12-31")
	companies := []Company{
		{BaseModel: shared.BaseModel{PublicID: "company_b", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-02")}, Name: "Beta", Slug: "beta", Status: CompanyStatusActive},
		{BaseModel: shared.BaseModel{PublicID: "company_a", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-01")}, Name: "Acme", Slug: "acme", Status: CompanyStatusActive},
		{BaseModel: shared.BaseModel{PublicID: "company_inactive", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-03")}, Name: "Inactive", Slug: "inactive", Status: CompanyStatusInactive},
		{BaseModel: shared.BaseModel{PublicID: "company_old", CreatedAt: mustCompanyRepositoryDate(t, "2024-03-01")}, Name: "Old Acme", Slug: "old-acme", Status: CompanyStatusActive},
	}
	if err := db.Create(&companies).Error; err != nil {
		t.Fatalf("create companies: %v", err)
	}
	repo := NewRepository(db)
	req := ListCompaniesRequest{ListParams: query.ListParams{
		Pagination: query.PaginationParams{Page: 1, PerPage: 10},
		Filters:    query.FilterParams{Status: CompanyStatusActive, CreatedFrom: &from, CreatedTo: &to},
		Sort:       query.SortParams{Sort: "name", Order: "asc"},
		Search:     query.SearchParams{Query: "a"},
	}}

	got, total, err := repo.List(req)
	if err != nil {
		t.Fatalf("list companies: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected filtered count 2, got %d", total)
	}
	if len(got) != 2 || got[0].PublicID != "company_a" || got[1].PublicID != "company_b" {
		t.Fatalf("expected active companies sorted by name, got %#v", got)
	}
}

func TestGormRepository_ListKeepsIncludeInactiveCompanySpecific(t *testing.T) {
	db := newCompanyRepositoryTestDB(t)
	companies := []Company{
		{BaseModel: shared.BaseModel{PublicID: "active", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-01")}, Name: "Active", Slug: "active", Status: CompanyStatusActive},
		{BaseModel: shared.BaseModel{PublicID: "inactive", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-02")}, Name: "Inactive", Slug: "inactive", Status: CompanyStatusInactive},
	}
	if err := db.Create(&companies).Error; err != nil {
		t.Fatalf("create companies: %v", err)
	}
	repo := NewRepository(db)

	activeOnly, total, err := repo.List(ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}}})
	if err != nil {
		t.Fatalf("list active companies: %v", err)
	}
	if total != 1 || len(activeOnly) != 1 || activeOnly[0].Status != CompanyStatusActive {
		t.Fatalf("expected inactive excluded by default, total=%d got=%#v", total, activeOnly)
	}

	all, total, err := repo.List(ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}}, IncludeInactive: true})
	if err != nil {
		t.Fatalf("list all companies: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected include_inactive to include both rows, total=%d got=%#v", total, all)
	}
}

func TestGormRepository_ListUsesExplicitSortAllowlist(t *testing.T) {
	db := newCompanyRepositoryTestDB(t)
	companies := []Company{
		{BaseModel: shared.BaseModel{PublicID: "new", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-02")}, Name: "Zeta", Slug: "zeta", Status: CompanyStatusActive},
		{BaseModel: shared.BaseModel{PublicID: "old", CreatedAt: mustCompanyRepositoryDate(t, "2025-03-01")}, Name: "Alpha", Slug: "alpha", Status: CompanyStatusActive},
	}
	if err := db.Create(&companies).Error; err != nil {
		t.Fatalf("create companies: %v", err)
	}
	repo := NewRepository(db)
	req := ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}, Sort: query.SortParams{Sort: "domain", Order: "asc"}}}

	got, _, err := repo.List(req)
	if err != nil {
		t.Fatalf("list companies: %v", err)
	}

	if len(got) != 2 || got[0].PublicID != "new" || got[1].PublicID != "old" {
		t.Fatalf("expected disallowed sort to fall back to created_at desc, got %#v", got)
	}
}

func TestGormRepository_DeleteSoftDeletesCompanies(t *testing.T) {
	db := newCompanyRepositoryTestDB(t)
	company := Company{BaseModel: shared.BaseModel{PublicID: "company_delete"}, Name: "Delete Me", Slug: "delete-me", Status: CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	repo := NewRepository(db)

	if err := repo.Delete("company_delete"); err != nil {
		t.Fatalf("delete company: %v", err)
	}
	if _, err := repo.GetByPublicID("company_delete"); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected normal read to exclude soft-deleted company, got %v", err)
	}
	var deleted Company
	if err := db.Unscoped().Where("public_id = ?", "company_delete").First(&deleted).Error; err != nil {
		t.Fatalf("expected soft-deleted company row to remain: %v", err)
	}
	if !deleted.DeletedAt.Valid {
		t.Fatal("expected deleted_at to be set")
	}
}

func newCompanyRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&Company{}); err != nil {
		t.Fatalf("migrate companies test db: %v", err)
	}
	return db
}

func mustCompanyRepositoryDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}
