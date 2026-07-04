package list_company_domains

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeListDomainsRepo struct {
	company *core.Company
	domains []core.CompanyDomain
	err     error
}

func (f *fakeListDomainsRepo) GetByPublicID(string) (*core.Company, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.company != nil {
		return f.company, nil
	}
	return &core.Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "cmp_01"}}, nil
}
func (f *fakeListDomainsRepo) ListDomains(uint) ([]core.CompanyDomain, error) {
	return f.domains, nil
}

func TestService_ListDomains(t *testing.T) {
	svc := NewService(&fakeListDomainsRepo{domains: []core.CompanyDomain{{BaseModel: shared.BaseModel{PublicID: "dom_01"}, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}}})
	got, err := svc.ListDomains("cmp_01")
	if err != nil || len(got) != 1 || got[0].CompanyPublicID != "cmp_01" {
		t.Fatalf("unexpected result: err=%v got=%#v", err, got)
	}
}

func TestService_ListDomains_NotFound(t *testing.T) {
	svc := NewService(&fakeListDomainsRepo{err: core.ErrCompanyNotFound})
	_, err := svc.ListDomains("missing")
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
