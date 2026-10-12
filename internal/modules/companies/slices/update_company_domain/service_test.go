package update_company_domain

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeUpdateDomainRepo struct {
	updated *company_domain.CompanyDomain
	err     error
}

func (f *fakeUpdateDomainRepo) GetByPublicID(string) (*company.Company, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &company.Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "cmp_01"}}, nil
}
func (f *fakeUpdateDomainRepo) GetDomainByPublicID(string) (*company_domain.CompanyDomain, error) {
	return &company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_01"}, CompanyID: 10, Domain: "old.com", Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusInactive}, nil
}
func (f *fakeUpdateDomainRepo) GetDomainByDomain(string) (*company_domain.CompanyDomain, error) {
	return nil, core.ErrCompanyDomainNotFound
}
func (f *fakeUpdateDomainRepo) CountActivePrimaryDomains(uint, string) (int64, error) { return 0, nil }
func (f *fakeUpdateDomainRepo) UpdateDomain(d *company_domain.CompanyDomain) error {
	f.updated = d
	return nil
}

func TestService_UpdateDomain(t *testing.T) {
	repo := &fakeUpdateDomainRepo{}
	svc := NewService(repo)
	got, err := svc.UpdateDomain("cmp_01", "dom_01", company_domain.UpdateCompanyDomainRequest{Domain: " Acme.com ", Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive, RedirectToPrimary: true})
	if err != nil {
		t.Fatal(err)
	}
	if got.Domain != "acme.com" || repo.updated == nil || repo.updated.Domain != "acme.com" || repo.updated.CompanyID != 10 || !repo.updated.RedirectToPrimary {
		t.Fatalf("unexpected result: err=%v got=%#v", err, got)
	}
}

func TestService_UpdateDomain_CompanyNotFound(t *testing.T) {
	repo := &fakeUpdateDomainRepo{err: core.ErrCompanyNotFound}
	svc := NewService(repo)

	_, err := svc.UpdateDomain("missing", "dom_01", company_domain.UpdateCompanyDomainRequest{Domain: "acme.com", Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive})
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
