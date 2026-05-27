package companies

import (
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
	if err := db.AutoMigrate(&core.Company{}, &core.CompanyDomain{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	company := core.Company{BaseModel: shared.BaseModel{PublicID: "company_acme"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	alias := core.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "d1"}, CompanyID: company.ID, Domain: "www.acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindAlias, RedirectToPrimary: true}
	primary := core.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "d2"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}
	if err := db.Create(&[]core.CompanyDomain{alias, primary}).Error; err != nil {
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
