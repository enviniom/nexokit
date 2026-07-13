package create_company_domain

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	internalshared "github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_QueryDelegation(t *testing.T) {
	// Query behavior is covered in companies/queries tests. This test verifies repository wiring/delegation semantics.
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{}, &core.CompanyDomain{})
	company := core.Company{BaseModel: internalshared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	_ = db.Create(&company).Error
	_ = db.Create(&core.CompanyDomain{BaseModel: internalshared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}).Error

	repo := NewRepository(db)
	if _, err := repo.GetByPublicID("cmp_01"); err != nil {
		t.Fatalf("GetByPublicID: %v", err)
	}
	if _, err := repo.GetDomainByDomain("acme.com"); err != nil {
		t.Fatalf("GetDomainByDomain: %v", err)
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

func TestGormRepository_GetDomainByDomain_NotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{}, &core.CompanyDomain{})
	company := core.Company{BaseModel: internalshared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	_ = db.Create(&company).Error

	repo := NewRepository(db)
	_, err := repo.GetDomainByDomain("missing.com")
	if !errors.Is(err, core.ErrCompanyDomainNotFound) {
		t.Fatalf("expected ErrCompanyDomainNotFound, got %v", err)
	}
}

func TestGormRepository_CreateDomain_UniqueViolation(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&core.Company{}, &core.CompanyDomain{})
	company := core.Company{BaseModel: internalshared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	_ = db.Create(&company).Error
	_ = db.Create(&core.CompanyDomain{BaseModel: internalshared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}).Error

	repo := NewRepository(db)
	domain := &core.CompanyDomain{BaseModel: internalshared.BaseModel{PublicID: "dom_02"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindAlias}
	err := repo.CreateDomain(domain)
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected ErrDuplicateCompanyDomain, got %v", err)
	}
}
