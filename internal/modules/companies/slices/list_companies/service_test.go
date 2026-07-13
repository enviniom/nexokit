package list_companies

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeListRepository struct {
	req   core.ListCompaniesRequest
	rows  []core.Company
	total int64
}

func (f *fakeListRepository) List(req core.ListCompaniesRequest) ([]core.Company, int64, error) {
	f.req = req
	return f.rows, f.total, nil
}

func TestService_List_DefaultPaginationAndExcludeInactive(t *testing.T) {
	repo := &fakeListRepository{rows: []core.Company{{BaseModel: shared.BaseModel{PublicID: "01H"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}}, total: 1}
	svc := NewService(repo)

	got, total, err := svc.List(core.ListCompaniesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.req.ListParams.Pagination.Page != 1 || repo.req.ListParams.Pagination.PerPage != 20 || repo.req.IncludeInactive {
		t.Fatalf("unexpected defaults: %#v", repo.req)
	}
	if total != 1 || len(got) != 1 || got[0].PublicID != "01H" {
		t.Fatalf("unexpected result: total=%d got=%#v", total, got)
	}
}

func TestService_List_PassesFiltersToRepository(t *testing.T) {
	repo := &fakeListRepository{rows: []core.Company{}, total: 0}
	svc := NewService(repo)
	req := core.ListCompaniesRequest{ListParams: query.ListParams{Pagination: query.PaginationParams{Page: 2, PerPage: 5}, Search: query.SearchParams{Query: "acme"}}, IncludeInactive: true}
	_, _, _ = svc.List(req)
	if repo.req.ListParams.Pagination.Page != 2 || repo.req.ListParams.Search.Query != "acme" || !repo.req.IncludeInactive {
		t.Fatalf("expected request pass-through, got %#v", repo.req)
	}
}
