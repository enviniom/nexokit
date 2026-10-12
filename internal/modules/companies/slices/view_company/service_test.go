package view_company

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeViewRepo struct {
	company *company.Company
	err     error
}

func (f *fakeViewRepo) GetByPublicID(string) (*company.Company, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.company, nil
}

func TestService_GetByPublicID_MapsDomains(t *testing.T) {
	repo := &fakeViewRepo{company: &company.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive, Domains: []company_domain.CompanyDomain{{BaseModel: shared.BaseModel{PublicID: "dom_01"}, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}}}}
	svc := NewService(repo)

	got, err := svc.GetByPublicID("cmp_01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicID != "cmp_01" || len(got.Domains) != 1 || got.Domains[0].Domain != "acme.com" {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestService_GetByPublicID_ReturnsNotFound(t *testing.T) {
	repo := &fakeViewRepo{err: core.ErrCompanyNotFound}
	svc := NewService(repo)

	_, err := svc.GetByPublicID("missing")
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
