package companies

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResolver_ResolveHostUsesActiveDomainAndCompany(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	company := company.Company{BaseModel: shared.BaseModel{PublicID: "company_acme"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	alias := company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "d1"}, CompanyID: company.ID, Domain: "www.acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindAlias, RedirectToPrimary: true}
	primary := company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "d2"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}
	if err := db.Create(&[]company_domain.CompanyDomain{alias, primary}).Error; err != nil {
		t.Fatalf("create domains: %v", err)
	}

	r := NewResolver(db)
	got, err := r.ResolveHost("www.acme.com")
	if err != nil {
		t.Fatalf("resolve host: %v", err)
	}
	if got.Company.Slug != "acme" || got.PrimaryDomain == nil || *got.PrimaryDomain != "acme.com" {
		t.Fatalf("unexpected host resolution: %#v", got)
	}
}
