package queries

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetCompanyByPublicID(t *testing.T) {
	db := newTestDB(t)
	seed := core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := GetCompanyByPublicID(db, "cmp_01")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got.PublicID != "cmp_01" {
		t.Fatalf("expected cmp_01, got %s", got.PublicID)
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
