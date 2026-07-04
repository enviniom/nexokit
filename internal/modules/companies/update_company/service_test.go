package update_company

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct {
	byID    map[string]*core.Company
	bySlug  map[string]*core.Company
	updated *core.Company
}

func (f *fakeRepo) GetByPublicID(publicID string) (*core.Company, error) {
	if c, ok := f.byID[publicID]; ok {
		return c, nil
	}
	return nil, core.ErrCompanyNotFound
}
func (f *fakeRepo) GetBySlugIncludingDeleted(slug string) (*core.Company, error) {
	if c, ok := f.bySlug[slug]; ok {
		return c, nil
	}
	return nil, core.ErrCompanyNotFound
}
func (f *fakeRepo) Update(c *core.Company) error { f.updated = c; return nil }

func TestService_Update_PreservesDomains(t *testing.T) {
	repo := &fakeRepo{byID: map[string]*core.Company{"01H": {BaseModel: shared.BaseModel{PublicID: "01H"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive, Domains: []core.CompanyDomain{{BaseModel: shared.BaseModel{PublicID: "D1"}, Domain: "acme.com"}}}}, bySlug: map[string]*core.Company{}}
	svc := NewService(repo)

	res, err := svc.Update("01H", core.UpdateCompanyRequest{Name: "Acme New", Slug: "acme", Status: core.CompanyStatusInactive})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != core.CompanyStatusInactive {
		t.Fatalf("expected inactive response, got %#v", res)
	}
	if repo.updated == nil || len(repo.updated.Domains) != 1 || repo.updated.Domains[0].PublicID != "D1" {
		t.Fatalf("expected update to preserve domains, got %#v", repo.updated)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := &fakeRepo{byID: map[string]*core.Company{}, bySlug: map[string]*core.Company{}}
	svc := NewService(repo)

	_, err := svc.Update("missing", core.UpdateCompanyRequest{Name: "X", Slug: "x", Status: core.CompanyStatusActive})
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}

func TestService_Update_DuplicateSlug(t *testing.T) {
	repo := &fakeRepo{byID: map[string]*core.Company{"01H": {BaseModel: shared.BaseModel{PublicID: "01H"}, Slug: "acme"}}, bySlug: map[string]*core.Company{"other": {BaseModel: shared.BaseModel{PublicID: "02H"}, Slug: "other"}}}
	svc := NewService(repo)

	_, err := svc.Update("01H", core.UpdateCompanyRequest{Name: "X", Slug: "other", Status: core.CompanyStatusActive})
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected duplicate slug, got %v", err)
	}
}
