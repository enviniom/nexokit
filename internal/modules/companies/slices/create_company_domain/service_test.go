package create_company_domain

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeDomainRepo struct {
	company     *company.Company
	byDomain    map[string]*company_domain.CompanyDomain
	activeCount int64
	created     *company_domain.CompanyDomain
}

func (f *fakeDomainRepo) GetByPublicID(string) (*company.Company, error) {
	if f.company == nil {
		return nil, core.ErrCompanyNotFound
	}
	return f.company, nil
}
func (f *fakeDomainRepo) GetDomainByDomain(d string) (*company_domain.CompanyDomain, error) {
	if v, ok := f.byDomain[d]; ok {
		return v, nil
	}
	return nil, core.ErrCompanyDomainNotFound
}
func (f *fakeDomainRepo) CountActivePrimaryDomains(uint, string) (int64, error) {
	return f.activeCount, nil
}
func (f *fakeDomainRepo) CreateDomain(d *company_domain.CompanyDomain) error {
	f.created = d
	return nil
}

func TestService_CreateDomain_DuplicateAndPrimaryConflict(t *testing.T) {
	repo := &fakeDomainRepo{company: &company.Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "01H"}}, byDomain: map[string]*company_domain.CompanyDomain{"www.acme.com": {BaseModel: shared.BaseModel{PublicID: "D1"}}}}
	svc := NewService(repo)
	_, err := svc.CreateDomain("01H", company_domain.CreateCompanyDomainRequest{Domain: "www.acme.com", Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected duplicate domain error, got %v", err)
	}

	repo.byDomain = map[string]*company_domain.CompanyDomain{}
	repo.activeCount = 1
	_, err = svc.CreateDomain("01H", company_domain.CreateCompanyDomainRequest{Domain: "primary.acme.com", Kind: core.CompanyDomainKindPrimary, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrActivePrimaryDomainExists) {
		t.Fatalf("expected active primary conflict, got %v", err)
	}
}

func TestService_CreateDomain_PersistsNormalizedDomain(t *testing.T) {
	repo := &fakeDomainRepo{company: &company.Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "cmp_01"}}}
	got, err := NewService(repo).CreateDomain("cmp_01", company_domain.CreateCompanyDomainRequest{Domain: "  ACME.COM  ", Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive, RedirectToPrimary: true})
	if err != nil {
		t.Fatal(err)
	}
	if repo.created == nil || repo.created.Domain != "acme.com" || repo.created.CompanyID != 10 || !repo.created.RedirectToPrimary || got.Domain != "acme.com" || got.CompanyPublicID != "cmp_01" {
		t.Fatalf("persisted=%+v response=%+v", repo.created, got)
	}
}

func TestService_CreateDomain_CompanyNotFound(t *testing.T) {
	repo := &fakeDomainRepo{}
	svc := NewService(repo)

	_, err := svc.CreateDomain("missing", company_domain.CreateCompanyDomainRequest{Domain: "acme.com", Kind: core.CompanyDomainKindPrimary, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
