package create_company_domain

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeDomainRepo struct {
	company     *core.Company
	byDomain    map[string]*core.CompanyDomain
	activeCount int64
	created     *core.CompanyDomain
}

func (f *fakeDomainRepo) GetByPublicID(string) (*core.Company, error) {
	if f.company == nil {
		return nil, core.ErrCompanyNotFound
	}
	return f.company, nil
}
func (f *fakeDomainRepo) GetDomainByDomain(d string) (*core.CompanyDomain, error) {
	if v, ok := f.byDomain[d]; ok {
		return v, nil
	}
	return nil, core.ErrCompanyDomainNotFound
}
func (f *fakeDomainRepo) CountActivePrimaryDomains(uint, string) (int64, error) {
	return f.activeCount, nil
}
func (f *fakeDomainRepo) CreateDomain(d *core.CompanyDomain) error {
	f.created = d
	return nil
}

func TestService_CreateDomain_DuplicateAndPrimaryConflict(t *testing.T) {
	repo := &fakeDomainRepo{company: &core.Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "01H"}}, byDomain: map[string]*core.CompanyDomain{"www.acme.com": {BaseModel: shared.BaseModel{PublicID: "D1"}}}}
	svc := NewService(repo)
	_, err := svc.CreateDomain("01H", core.CreateCompanyDomainRequest{Domain: "www.acme.com", Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected duplicate domain error, got %v", err)
	}

	repo.byDomain = map[string]*core.CompanyDomain{}
	repo.activeCount = 1
	_, err = svc.CreateDomain("01H", core.CreateCompanyDomainRequest{Domain: "primary.acme.com", Kind: core.CompanyDomainKindPrimary, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrActivePrimaryDomainExists) {
		t.Fatalf("expected active primary conflict, got %v", err)
	}
}

func TestService_CreateDomain_CompanyNotFound(t *testing.T) {
	repo := &fakeDomainRepo{}
	svc := NewService(repo)

	_, err := svc.CreateDomain("missing", core.CreateCompanyDomainRequest{Domain: "acme.com", Kind: core.CompanyDomainKindPrimary, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
