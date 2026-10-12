package update_company_domain

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_QueryDelegation(t *testing.T) {
	// Query behavior is covered in companies/queries tests. This test verifies repository wiring/delegation semantics.
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{})
	company := company.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	_ = db.Create(&company).Error
	_ = db.Create(&company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}).Error

	repo := NewRepository(db)
	if _, err := repo.GetByPublicID("cmp_01"); err != nil {
		t.Fatalf("GetByPublicID: %v", err)
	}
	if _, err := repo.GetDomainByDomain("acme.com"); err != nil {
		t.Fatalf("GetDomainByDomain: %v", err)
	}
}

func TestGormRepository_GetDomainByPublicID_NotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{})
	company := company.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	_ = db.Create(&company).Error

	repo := NewRepository(db)
	_, err := repo.GetDomainByPublicID("missing")
	if !errors.Is(err, core.ErrCompanyDomainNotFound) {
		t.Fatalf("expected ErrCompanyDomainNotFound, got %v", err)
	}
}

func TestGormRepository_UpdateDomain_UniqueViolation(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{})
	company := company.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	_ = db.Create(&company).Error
	_ = db.Create(&company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}).Error
	_ = db.Create(&company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_02"}, CompanyID: company.ID, Domain: "www.acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindAlias}).Error

	repo := NewRepository(db)
	domain := &company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_02"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindAlias}
	err := repo.UpdateDomain(domain)
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected ErrDuplicateCompanyDomain, got %v", err)
	}
}

func TestGormRepository_UpdateDomain_ZeroRowsIsNotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{})
	err := NewRepository(db).UpdateDomain(&company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "missing"}, Domain: "missing.example", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindAlias})
	if !errors.Is(err, core.ErrCompanyDomainNotFound) {
		t.Fatalf("expected ErrCompanyDomainNotFound, got %v", err)
	}
}
