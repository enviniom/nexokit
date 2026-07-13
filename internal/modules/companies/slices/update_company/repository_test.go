package update_company

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_GetByPublicID_DelegatesToQueries(t *testing.T) {
	// Query behavior is covered in companies/queries tests. This test verifies repository wiring/delegation semantics.
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{})
	_ = db.Create(&core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}).Error

	repo := NewRepository(db)
	got, err := repo.GetByPublicID("cmp_01")
	if err != nil || got.PublicID != "cmp_01" {
		t.Fatalf("expected delegated company lookup, got err=%v company=%#v", err, got)
	}
}

func TestGormRepository_GetByPublicID_NotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{})

	repo := NewRepository(db)
	_, err := repo.GetByPublicID("missing")
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}

func TestGormRepository_Update_UniqueViolation(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{})
	_ = db.Create(&core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_02"}, Name: "Other", Slug: "other", Status: core.CompanyStatusActive}).Error

	repo := NewRepository(db)
	company := &core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_02"}, Name: "Other", Slug: "acme", Status: core.CompanyStatusActive}
	err := repo.Update(company)
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected ErrDuplicateCompanySlug, got %v", err)
	}
}

func TestGormRepository_Update_ZeroRowsIsNotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{})
	err := NewRepository(db).Update(&core.Company{BaseModel: shared.BaseModel{PublicID: "missing"}, Name: "Missing", Slug: "missing", Status: core.CompanyStatusActive})
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
