package companies

import (
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type fakeCompanyRepository struct {
	companies                 []Company
	domains                   []CompanyDomain
	byPublicID                map[string]*Company
	bySlug                    map[string]*Company
	byDomainPublicID          map[string]*CompanyDomain
	byDomain                  map[string]*CompanyDomain
	total                     int64
	activePrimaryCount        int64
	err                       error
	created                   *Company
	updated                   *Company
	createdDomain             *CompanyDomain
	updatedDomain             *CompanyDomain
	deletedPublicID           string
	listReq                   ListCompaniesRequest
	countActivePrimaryCompany uint
	countActivePrimaryExclude string
}

func (f *fakeCompanyRepository) List(req ListCompaniesRequest) ([]Company, int64, error) {
	f.listReq = req
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.companies, f.total, nil
}

func (f *fakeCompanyRepository) GetByPublicID(publicID string) (*Company, error) {
	if f.err != nil {
		return nil, f.err
	}
	if company, ok := f.byPublicID[publicID]; ok {
		return company, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeCompanyRepository) GetBySlugIncludingDeleted(slug string) (*Company, error) {
	if f.err != nil {
		return nil, f.err
	}
	if company, ok := f.bySlug[slug]; ok {
		return company, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeCompanyRepository) Create(company *Company) error {
	if f.err != nil {
		return f.err
	}
	f.created = company
	return nil
}

func (f *fakeCompanyRepository) Update(company *Company) error {
	if f.err != nil {
		return f.err
	}
	f.updated = company
	return nil
}

func (f *fakeCompanyRepository) Delete(publicID string) error {
	f.deletedPublicID = publicID
	return f.err
}

func (f *fakeCompanyRepository) ListDomains(companyID uint) ([]CompanyDomain, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.domains, nil
}

func (f *fakeCompanyRepository) GetDomainByPublicID(publicID string) (*CompanyDomain, error) {
	if f.err != nil {
		return nil, f.err
	}
	if domain, ok := f.byDomainPublicID[publicID]; ok {
		return domain, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeCompanyRepository) GetDomainByDomain(domain string) (*CompanyDomain, error) {
	if f.err != nil {
		return nil, f.err
	}
	if domainRow, ok := f.byDomain[domain]; ok {
		return domainRow, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeCompanyRepository) CountActivePrimaryDomains(companyID uint, excludePublicID string) (int64, error) {
	f.countActivePrimaryCompany = companyID
	f.countActivePrimaryExclude = excludePublicID
	return f.activePrimaryCount, f.err
}

func (f *fakeCompanyRepository) CreateDomain(domain *CompanyDomain) error {
	if f.err != nil {
		return f.err
	}
	f.createdDomain = domain
	return nil
}

func (f *fakeCompanyRepository) UpdateDomain(domain *CompanyDomain) error {
	if f.err != nil {
		return f.err
	}
	f.updatedDomain = domain
	return nil
}

func TestService_Create(t *testing.T) {
	t.Run("root creates a company with generated public id", func(t *testing.T) {
		repo := &fakeCompanyRepository{bySlug: map[string]*Company{}}
		svc := NewService(repo)

		result, err := svc.Create(CreateCompanyRequest{Name: "Acme", Slug: "acme"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.PublicID == "" {
			t.Fatal("expected generated public_id")
		}
		if result.Name != "Acme" || result.Slug != "acme" {
			t.Fatalf("expected created Acme company, got %#v", result)
		}
		if repo.created == nil || repo.created.Status != CompanyStatusActive {
			t.Fatalf("expected active company persisted, got %#v", repo.created)
		}
	})

	t.Run("duplicate slug is rejected before persistence", func(t *testing.T) {
		repo := &fakeCompanyRepository{bySlug: map[string]*Company{"acme": {BaseModel: shared.BaseModel{PublicID: "01HACME"}, Slug: "acme"}}}
		svc := NewService(repo)

		_, err := svc.Create(CreateCompanyRequest{Name: "Acme 2", Slug: "acme"})
		if !errors.Is(err, ErrDuplicateSlug) {
			t.Fatalf("expected ErrDuplicateSlug, got %v", err)
		}
		if repo.created != nil {
			t.Fatal("duplicate slug must not be persisted")
		}
	})
}

func TestService_List(t *testing.T) {
	t.Run("excludes inactive companies by default", func(t *testing.T) {
		repo := &fakeCompanyRepository{
			companies: []Company{{BaseModel: shared.BaseModel{PublicID: "01HACTIVE"}, Name: "Active", Slug: "active", Status: CompanyStatusActive}},
			total:     1,
		}
		svc := NewService(repo)

		result, total, err := svc.List(ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.Pagination{Page: 1, PerPage: 20}}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.listReq.IncludeInactive {
			t.Fatal("expected inactive companies to be excluded by default")
		}
		if total != 1 || len(result) != 1 || result[0].Status != CompanyStatusActive {
			t.Fatalf("expected one active company, got total=%d result=%#v", total, result)
		}
	})

	t.Run("includes inactive companies when explicitly requested", func(t *testing.T) {
		repo := &fakeCompanyRepository{
			companies: []Company{{BaseModel: shared.BaseModel{PublicID: "01HINACTIVE"}, Name: "Inactive", Slug: "inactive", Status: CompanyStatusInactive}},
			total:     1,
		}
		svc := NewService(repo)

		result, _, err := svc.List(ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.Pagination{Page: 1, PerPage: 20}}, IncludeInactive: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !repo.listReq.IncludeInactive {
			t.Fatal("expected include inactive filter to reach repository")
		}
		if len(result) != 1 || result[0].Status != CompanyStatusInactive {
			t.Fatalf("expected inactive company, got %#v", result)
		}
	})

	t.Run("passes shared filters search sorting and dates to repository", func(t *testing.T) {
		repo := &fakeCompanyRepository{
			companies: []Company{{BaseModel: shared.BaseModel{PublicID: "01HACME"}, Name: "Acme", Slug: "acme", Status: CompanyStatusActive}},
			total:     1,
		}
		svc := NewService(repo)
		from := mustCompanyDate(t, "2025-01-01")
		to := mustCompanyDate(t, "2025-12-31")
		req := ListCompaniesRequest{
			ListParams: query.ListParams{
				Pagination: query.PaginationParams{Page: 2, PerPage: 5},
				Filters:    query.FilterParams{Status: CompanyStatusActive, CreatedFrom: &from, CreatedTo: &to},
				Sort:       query.SortParams{Sort: "slug", Order: "asc"},
				Search:     query.SearchParams{Query: "acme"},
			},
			IncludeInactive: true,
		}

		result, total, err := svc.List(req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 || len(result) != 1 || result[0].Slug != "acme" {
			t.Fatalf("expected Acme result, total=%d result=%#v", total, result)
		}
		if !repo.listReq.IncludeInactive || repo.listReq.ListParams.Sort.Sort != "slug" || repo.listReq.ListParams.Search.Query != "acme" || repo.listReq.ListParams.Filters.Status != CompanyStatusActive {
			t.Fatalf("expected repository to receive shared and company filters, got %#v", repo.listReq)
		}
	})
}

func mustCompanyDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}

func TestService_GetUpdateDeleteUsePublicID(t *testing.T) {
	repo := &fakeCompanyRepository{
		byPublicID: map[string]*Company{"01HCOMPANYPUBLICID": {
			BaseModel: shared.BaseModel{PublicID: "01HCOMPANYPUBLICID"},
			Name:      "Acme",
			Slug:      "acme",
			Status:    CompanyStatusActive,
			Domains: []CompanyDomain{
				{BaseModel: shared.BaseModel{PublicID: "01HDOMAIN"}, Domain: "acme.com", Kind: CompanyDomainKindPrimary, Status: CompanyDomainStatusActive},
			},
		}},
		bySlug: map[string]*Company{},
	}
	svc := NewService(repo)

	got, err := svc.GetByPublicID("01HCOMPANYPUBLICID")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got.PublicID != "01HCOMPANYPUBLICID" {
		t.Fatalf("expected public id response, got %#v", got)
	}
	if len(got.Domains) != 1 || got.Domains[0].Domain != "acme.com" {
		t.Fatalf("expected detail response to include domains, got %#v", got.Domains)
	}

	updated, err := svc.Update("01HCOMPANYPUBLICID", UpdateCompanyRequest{Name: "Acme Updated", Slug: "acme", Status: CompanyStatusInactive})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.Status != CompanyStatusInactive || repo.updated.PublicID != "01HCOMPANYPUBLICID" {
		t.Fatalf("expected update by public id, got updated=%#v repo=%#v", updated, repo.updated)
	}

	if err := svc.Delete("01HCOMPANYPUBLICID"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if repo.deletedPublicID != "01HCOMPANYPUBLICID" {
		t.Fatalf("expected delete by public id, got %q", repo.deletedPublicID)
	}
}

func TestService_GetReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeCompanyRepository{byPublicID: map[string]*Company{}})

	_, err := svc.GetByPublicID("missing")
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestService_CompanyDomainAdministration(t *testing.T) {
	company := &Company{BaseModel: shared.BaseModel{ID: 10, PublicID: "01HCOMPANY"}, Name: "Acme", Slug: "acme", Status: CompanyStatusActive}

	t.Run("lists multiple domains for company public id", func(t *testing.T) {
		repo := &fakeCompanyRepository{
			byPublicID: map[string]*Company{"01HCOMPANY": company},
			domains: []CompanyDomain{
				{BaseModel: shared.BaseModel{PublicID: "01HPRIMARY"}, CompanyID: company.ID, Domain: "acme.com", Kind: CompanyDomainKindPrimary, Status: CompanyDomainStatusActive},
				{BaseModel: shared.BaseModel{PublicID: "01HALIAS"}, CompanyID: company.ID, Domain: "www.acme.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusInactive},
			},
		}
		svc := NewService(repo)

		got, err := svc.ListDomains("01HCOMPANY")
		if err != nil {
			t.Fatalf("unexpected list domains error: %v", err)
		}
		if len(got) != 2 || got[0].CompanyPublicID != "01HCOMPANY" || got[1].Domain != "www.acme.com" {
			t.Fatalf("expected company domains response, got %#v", got)
		}
	})

	t.Run("creates normalized alias domain", func(t *testing.T) {
		repo := &fakeCompanyRepository{byPublicID: map[string]*Company{"01HCOMPANY": company}, byDomain: map[string]*CompanyDomain{}}
		svc := NewService(repo)

		got, err := svc.CreateDomain("01HCOMPANY", CreateCompanyDomainRequest{Domain: " WWW.Acme.COM. ", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive, RedirectToPrimary: true})
		if err != nil {
			t.Fatalf("unexpected create domain error: %v", err)
		}
		if got.Domain != "www.acme.com" || !got.RedirectToPrimary || repo.createdDomain.CompanyID != company.ID {
			t.Fatalf("expected normalized created domain, got response=%#v stored=%#v", got, repo.createdDomain)
		}
	})

	t.Run("rejects duplicate domain globally", func(t *testing.T) {
		repo := &fakeCompanyRepository{byPublicID: map[string]*Company{"01HCOMPANY": company}, byDomain: map[string]*CompanyDomain{"www.acme.com": {BaseModel: shared.BaseModel{PublicID: "01HEXISTING"}, Domain: "www.acme.com"}}}
		svc := NewService(repo)

		_, err := svc.CreateDomain("01HCOMPANY", CreateCompanyDomainRequest{Domain: "www.acme.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive})
		if !errors.Is(err, ErrDuplicateCompanyDomain) {
			t.Fatalf("expected duplicate domain error, got %v", err)
		}
	})

	t.Run("domain validation rejects normalized empty values", func(t *testing.T) {
		createErrs := CreateCompanyDomainRequest{Domain: "   ", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive}.Validate()
		if len(createErrs) == 0 {
			t.Fatal("expected whitespace-only create domain to fail validation")
		}
		updateErrs := UpdateCompanyDomainRequest{Domain: ".", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive}.Validate()
		if len(updateErrs) == 0 {
			t.Fatal("expected normalized-empty update domain to fail validation")
		}
	})

	t.Run("rejects second active primary", func(t *testing.T) {
		repo := &fakeCompanyRepository{byPublicID: map[string]*Company{"01HCOMPANY": company}, byDomain: map[string]*CompanyDomain{}, activePrimaryCount: 1}
		svc := NewService(repo)

		_, err := svc.CreateDomain("01HCOMPANY", CreateCompanyDomainRequest{Domain: "primary.acme.com", Kind: CompanyDomainKindPrimary, Status: CompanyDomainStatusActive})
		if !errors.Is(err, ErrActivePrimaryDomainExists) {
			t.Fatalf("expected active primary conflict, got %v", err)
		}
	})

	t.Run("updates only domains that belong to the specified company", func(t *testing.T) {
		repo := &fakeCompanyRepository{
			byPublicID:       map[string]*Company{"01HCOMPANY": company},
			byDomain:         map[string]*CompanyDomain{},
			byDomainPublicID: map[string]*CompanyDomain{"01HDOMAIN": {BaseModel: shared.BaseModel{PublicID: "01HDOMAIN"}, CompanyID: 999, Domain: "other.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive}},
		}
		svc := NewService(repo)

		_, err := svc.UpdateDomain("01HCOMPANY", "01HDOMAIN", UpdateCompanyDomainRequest{Domain: "other.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusInactive})
		if !errors.Is(err, ErrCompanyDomainDoesNotBelong) {
			t.Fatalf("expected ownership error, got %v", err)
		}
	})
}
