package list_company_domains

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeListDomainsRepo struct{}

func (f *fakeListDomainsRepo) GetByPublicID(string) (*core.Company, error) {
	return &core.Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "cmp_01"}}, nil
}
func (f *fakeListDomainsRepo) ListDomains(uint) ([]core.CompanyDomain, error) {
	return []core.CompanyDomain{{BaseModel: shared.BaseModel{PublicID: "dom_01"}, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}}, nil
}

func TestService_ListDomains(t *testing.T) {
	svc := NewService(&fakeListDomainsRepo{})
	got, err := svc.ListDomains("cmp_01")
	if err != nil || len(got) != 1 || got[0].CompanyPublicID != "cmp_01" {
		t.Fatalf("unexpected result: err=%v got=%#v", err, got)
	}
}
