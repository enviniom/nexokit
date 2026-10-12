package list_company_domains

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeListDomainsService struct{}

func (f *fakeListDomainsService) ListDomains(string) ([]company_domain.CompanyDomainResponse, error) {
	return []company_domain.CompanyDomainResponse{{PublicID: "dom_01", Domain: "acme.com"}}, nil
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/companies/:id/domains", NewHandler(&fakeListDomainsService{}).Handle)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/companies/cmp_01/domains", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
