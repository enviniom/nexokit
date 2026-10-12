package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestGetCompanyDomainByDomain(t *testing.T) {
	db := newTestDB(t)
	company := company.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	seed := company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed domain: %v", err)
	}

	got, err := GetCompanyDomainByDomain(db, "acme.com")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got.PublicID != "dom_01" {
		t.Fatalf("expected dom_01, got %s", got.PublicID)
	}
}
