package list_companies

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeListService struct {
	req   core.ListCompaniesRequest
	data  []core.CompanyResponse
	total int64
	err   error
}

func (f *fakeListService) List(req core.ListCompaniesRequest) ([]core.CompanyResponse, int64, error) {
	f.req = req
	return f.data, f.total, f.err
}

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeListService{data: []core.CompanyResponse{{PublicID: "01H", Status: core.CompanyStatusActive}}, total: 1}
	r := gin.New()
	r.GET("/companies", NewHandler(svc).Handle)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/companies?page=2&per_page=5&include_inactive=true&status=active&search=acme", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !svc.req.IncludeInactive || svc.req.ListParams.Pagination.Page != 2 || svc.req.ListParams.Pagination.PerPage != 5 || svc.req.ListParams.Filters.Status != "active" || svc.req.ListParams.Search.Query != "acme" {
		t.Fatalf("unexpected request parsed: %#v", svc.req)
	}
	var resp response.APIResponse[[]core.CompanyResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected one company in response, got %#v", resp.Data)
	}
}
