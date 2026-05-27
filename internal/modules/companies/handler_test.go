package companies

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeCompanyService struct {
	companies       []CompanyResponse
	company         *CompanyResponse
	domains         []CompanyDomainResponse
	domain          *CompanyDomainResponse
	total           int64
	err             error
	lastID          string
	lastCompanyID   string
	lastDomainID    string
	listReq         ListCompaniesRequest
	createDomainReq CreateCompanyDomainRequest
	updateDomainReq UpdateCompanyDomainRequest
}

func (f *fakeCompanyService) List(req ListCompaniesRequest) ([]CompanyResponse, int64, error) {
	f.listReq = req
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.companies, f.total, nil
}

func (f *fakeCompanyService) GetByPublicID(publicID string) (*CompanyResponse, error) {
	f.lastID = publicID
	if f.err != nil {
		return nil, f.err
	}
	return f.company, nil
}

func (f *fakeCompanyService) Create(req CreateCompanyRequest) (*CompanyResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.company, nil
}

func (f *fakeCompanyService) Update(publicID string, req UpdateCompanyRequest) (*CompanyResponse, error) {
	f.lastID = publicID
	if f.err != nil {
		return nil, f.err
	}
	return f.company, nil
}

func (f *fakeCompanyService) Delete(publicID string) error {
	f.lastID = publicID
	return f.err
}

func (f *fakeCompanyService) ListDomains(companyPublicID string) ([]CompanyDomainResponse, error) {
	f.lastCompanyID = companyPublicID
	if f.err != nil {
		return nil, f.err
	}
	return f.domains, nil
}

func (f *fakeCompanyService) CreateDomain(companyPublicID string, req CreateCompanyDomainRequest) (*CompanyDomainResponse, error) {
	f.lastCompanyID = companyPublicID
	f.createDomainReq = req
	if f.err != nil {
		return nil, f.err
	}
	return f.domain, nil
}

func (f *fakeCompanyService) UpdateDomain(companyPublicID, domainPublicID string, req UpdateCompanyDomainRequest) (*CompanyDomainResponse, error) {
	f.lastCompanyID = companyPublicID
	f.lastDomainID = domainPublicID
	f.updateDomainReq = req
	if f.err != nil {
		return nil, f.err
	}
	return f.domain, nil
}

func companyJSONRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func setupCompanyRouter(user *authctx.User, svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if user != nil {
			authctx.SetGin(c, user)
		}
		c.Next()
	})
	Register(r.Group("/api/v1"), NewHandler(svc), middleware.RequirePermission, middleware.RequireRole)
	return r
}

func TestHandler_Create(t *testing.T) {
	t.Run("direct company creation is blocked for all users (returns 404)", func(t *testing.T) {
		for _, role := range []string{"root", "admin", "user"} {
			t.Run(role, func(t *testing.T) {
				svc := &fakeCompanyService{company: &CompanyResponse{PublicID: "01HCOMPANY", Name: "Acme", Slug: "acme"}}
				isRoot := role == "root"
				r := setupCompanyRouter(&authctx.User{RoleSlug: role, IsRoot: isRoot}, svc)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, companyJSONRequest(http.MethodPost, "/api/v1/companies", CreateCompanyRequest{Name: "Acme", Slug: "acme"}))

				if w.Code != http.StatusNotFound {
					t.Fatalf("expected status 404, got %d", w.Code)
				}
			})
		}
	})
}

func TestHandler_ListFiltersInactive(t *testing.T) {
	t.Run("defaults to excluding inactive companies", func(t *testing.T) {
		svc := &fakeCompanyService{companies: []CompanyResponse{{PublicID: "01HACTIVE", Name: "Active", Slug: "active", Status: CompanyStatusActive}}, total: 1}
		r := setupCompanyRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/companies", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if svc.listReq.IncludeInactive {
			t.Fatal("expected inactive companies excluded by default")
		}
		var resp response.APIResponse[[]CompanyResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(resp.Data) != 1 || resp.Data[0].Status != CompanyStatusActive {
			t.Fatalf("expected active company only, got %#v", resp.Data)
		}
	})

	t.Run("passes shared list params and returns filter metadata", func(t *testing.T) {
		svc := &fakeCompanyService{companies: []CompanyResponse{{PublicID: "01HACME", Name: "Acme", Slug: "acme", Status: CompanyStatusActive}}, total: 1}
		r := setupCompanyRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/companies?page=2&per_page=5&status=active&created_from=2025-01-01&created_to=2025-12-31&sort=name&order=asc&search=acme&include_inactive=true", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
		}
		if !svc.listReq.IncludeInactive {
			t.Fatal("expected include_inactive to remain company-specific")
		}
		if svc.listReq.ListParams.Pagination.Page != 2 || svc.listReq.ListParams.Pagination.PerPage != 5 {
			t.Fatalf("expected parsed pagination, got %#v", svc.listReq.ListParams.Pagination)
		}
		if svc.listReq.ListParams.Filters.Status != "active" || svc.listReq.ListParams.Filters.CreatedFrom == nil || svc.listReq.ListParams.Filters.CreatedTo == nil {
			t.Fatalf("expected status/date filters, got %#v", svc.listReq.ListParams.Filters)
		}
		if svc.listReq.ListParams.Sort.Sort != "name" || svc.listReq.ListParams.Sort.Order != "asc" || svc.listReq.ListParams.Search.Query != "acme" {
			t.Fatalf("expected sort/search params, got sort=%#v search=%#v", svc.listReq.ListParams.Sort, svc.listReq.ListParams.Search)
		}

		var resp response.APIResponse[[]CompanyResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		filters := resp.Meta.(map[string]any)["filters"].(map[string]any)
		if filters["status"] != "active" || filters["sort"] != "name" || filters["order"] != "asc" || filters["search"] != "acme" {
			t.Fatalf("expected filter metadata to reflect query, got %#v", filters)
		}
	})
}

func TestHandler_UsesPublicIDRoutes(t *testing.T) {
	svc := &fakeCompanyService{company: &CompanyResponse{PublicID: "01HCOMPANYPUBLICID", Name: "Acme", Slug: "acme", Status: CompanyStatusActive, Domains: []CompanyDomainResponse{{PublicID: "01HDOMAIN", Domain: "acme.com", Kind: CompanyDomainKindPrimary, Status: CompanyDomainStatusActive}}}}
	r := setupCompanyRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

	for _, tt := range []struct {
		name       string
		method     string
		path       string
		body       any
		wantStatus int
	}{
		{name: "get", method: http.MethodGet, path: "/api/v1/companies/01HCOMPANYPUBLICID", wantStatus: http.StatusOK},
		{name: "update", method: http.MethodPut, path: "/api/v1/companies/01HCOMPANYPUBLICID", body: UpdateCompanyRequest{Name: "Acme", Slug: "acme", Status: CompanyStatusActive}, wantStatus: http.StatusOK},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/companies/01HCOMPANYPUBLICID", wantStatus: http.StatusNoContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, companyJSONRequest(tt.method, tt.path, tt.body))
			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tt.wantStatus, w.Code, w.Body.String())
			}
			if tt.method == http.MethodGet {
				var resp response.APIResponse[CompanyResponse]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("decode get response: %v", err)
				}
				if len(resp.Data.Domains) != 1 || resp.Data.Domains[0].Domain != "acme.com" {
					t.Fatalf("expected detail response to include domains, got %#v", resp.Data.Domains)
				}
			}
			if tt.method == http.MethodDelete && w.Body.Len() != 0 {
				t.Fatalf("expected empty body, got %q", w.Body.String())
			}
			if svc.lastID != "01HCOMPANYPUBLICID" {
				t.Fatalf("expected public id route param, got %q", svc.lastID)
			}
		})
	}
}

func TestHandler_NotFound(t *testing.T) {
	svc := &fakeCompanyService{err: apperror.ErrNotFound}
	r := setupCompanyRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/companies/missing", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandler_CompanyDomainsRootAdmin(t *testing.T) {
	t.Run("list create and update use nested public ids", func(t *testing.T) {
		svc := &fakeCompanyService{
			domains: []CompanyDomainResponse{{PublicID: "01HDOMAIN", Domain: "acme.com", Kind: CompanyDomainKindPrimary, Status: CompanyDomainStatusActive}},
			domain:  &CompanyDomainResponse{PublicID: "01HDOMAIN", Domain: "www.acme.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive, RedirectToPrimary: true},
		}
		r := setupCompanyRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/companies/01HCOMPANY/domains", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected list status 200, got %d body=%s", w.Code, w.Body.String())
		}
		if svc.lastCompanyID != "01HCOMPANY" {
			t.Fatalf("expected company public id route param, got %q", svc.lastCompanyID)
		}

		w = httptest.NewRecorder()
		r.ServeHTTP(w, companyJSONRequest(http.MethodPost, "/api/v1/companies/01HCOMPANY/domains", CreateCompanyDomainRequest{Domain: "www.acme.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusActive, RedirectToPrimary: true}))
		if w.Code != http.StatusCreated {
			t.Fatalf("expected create status 201, got %d body=%s", w.Code, w.Body.String())
		}
		if svc.createDomainReq.Domain != "www.acme.com" || !svc.createDomainReq.RedirectToPrimary {
			t.Fatalf("expected create domain request to reach service, got %#v", svc.createDomainReq)
		}

		w = httptest.NewRecorder()
		r.ServeHTTP(w, companyJSONRequest(http.MethodPut, "/api/v1/companies/01HCOMPANY/domains/01HDOMAIN", UpdateCompanyDomainRequest{Domain: "www.acme.com", Kind: CompanyDomainKindAlias, Status: CompanyDomainStatusInactive}))
		if w.Code != http.StatusOK {
			t.Fatalf("expected update status 200, got %d body=%s", w.Code, w.Body.String())
		}
		if svc.lastCompanyID != "01HCOMPANY" || svc.lastDomainID != "01HDOMAIN" || svc.updateDomainReq.Status != CompanyDomainStatusInactive {
			t.Fatalf("expected nested update params/request, company=%q domain=%q req=%#v", svc.lastCompanyID, svc.lastDomainID, svc.updateDomainReq)
		}
	})

	t.Run("non-root cannot administer domains and delete route is absent", func(t *testing.T) {
		svc := &fakeCompanyService{}
		r := setupCompanyRouter(&authctx.User{RoleSlug: "admin", IsRoot: false}, svc)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/companies/01HCOMPANY/domains", nil))
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected non-root list status 403, got %d", w.Code)
		}

		w = httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/companies/01HCOMPANY/domains/01HDOMAIN", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected delete route 404, got %d", w.Code)
		}
	})
}
