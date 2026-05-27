package list_companies

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_ListFiltersAndIncludeInactive(t *testing.T) {
	db := newTestDB(t)
	rows := []core.Company{
		{BaseModel: shared.BaseModel{PublicID: "a", CreatedAt: mustDate(t, "2025-03-01")}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive},
		{BaseModel: shared.BaseModel{PublicID: "i", CreatedAt: mustDate(t, "2025-03-02")}, Name: "Inactive", Slug: "inactive", Status: core.CompanyStatusInactive},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := NewRepository(db)

	active, total, err := repo.List(core.ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}}})
	if err != nil || total != 1 || len(active) != 1 || active[0].Status != core.CompanyStatusActive {
		t.Fatalf("expected active only, total=%d len=%d err=%v", total, len(active), err)
	}

	all, total, err := repo.List(core.ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}, Search: query.SearchParams{Query: "a"}}, IncludeInactive: true})
	if err != nil || total != 2 || len(all) != 2 {
		t.Fatalf("expected include inactive list, total=%d len=%d err=%v", total, len(all), err)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.Company{}, &core.CompanyDomain{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	v, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return v
}
