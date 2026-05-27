package queries

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestCountActivePrimaryDomains(t *testing.T) {
	db := newTestDB(t)
	company := core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	rows := []core.CompanyDomain{
		{BaseModel: shared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "a.acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary},
		{BaseModel: shared.BaseModel{PublicID: "dom_02"}, CompanyID: company.ID, Domain: "b.acme.com", Status: core.CompanyDomainStatusInactive, Kind: core.CompanyDomainKindPrimary},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed domains: %v", err)
	}

	n, err := CountActivePrimaryDomains(db, company.ID, "")
	if err != nil || n != 1 {
		t.Fatalf("expected 1 active primary, got n=%d err=%v", n, err)
	}
	n, err = CountActivePrimaryDomains(db, company.ID, "dom_01")
	if err != nil || n != 0 {
		t.Fatalf("expected 0 after exclude, got n=%d err=%v", n, err)
	}
}
