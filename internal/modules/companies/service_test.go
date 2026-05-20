package companies

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type fakeCompanyRepository struct {
	companies       []Company
	byPublicID      map[string]*Company
	bySlug          map[string]*Company
	total           int64
	err             error
	created         *Company
	updated         *Company
	deletedPublicID string
	listReq         ListCompaniesRequest
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

		result, total, err := svc.List(ListCompaniesRequest{Pagination: query.Pagination{Page: 1, PerPage: 20}})
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

		result, _, err := svc.List(ListCompaniesRequest{Pagination: query.Pagination{Page: 1, PerPage: 20}, IncludeInactive: true})
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
}

func TestService_GetUpdateDeleteUsePublicID(t *testing.T) {
	repo := &fakeCompanyRepository{
		byPublicID: map[string]*Company{"01HCOMPANYPUBLICID": {BaseModel: shared.BaseModel{PublicID: "01HCOMPANYPUBLICID"}, Name: "Acme", Slug: "acme", Status: CompanyStatusActive}},
		bySlug:     map[string]*Company{},
	}
	svc := NewService(repo)

	got, err := svc.GetByPublicID("01HCOMPANYPUBLICID")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got.PublicID != "01HCOMPANYPUBLICID" {
		t.Fatalf("expected public id response, got %#v", got)
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
