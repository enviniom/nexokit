package view_company

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeViewRepo struct{ company *core.Company }

func (f *fakeViewRepo) GetByPublicID(string) (*core.Company, error) { return f.company, nil }

func TestService_GetByPublicID_MapsDomains(t *testing.T) {
	repo := &fakeViewRepo{company: &core.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive, Domains: []core.CompanyDomain{{BaseModel: shared.BaseModel{PublicID: "dom_01"}, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}}}}
	svc := NewService(repo)

	got, err := svc.GetByPublicID("cmp_01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicID != "cmp_01" || len(got.Domains) != 1 || got.Domains[0].Domain != "acme.com" {
		t.Fatalf("unexpected response: %#v", got)
	}
}
