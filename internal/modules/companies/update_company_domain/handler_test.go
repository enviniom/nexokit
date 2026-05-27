package update_company_domain

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/gin-gonic/gin"
)

type fakeUpdateDomainService struct{}

func (f *fakeUpdateDomainService) UpdateDomain(string, string, core.UpdateCompanyDomainRequest) (*core.CompanyDomainResponse, error) {
	return &core.CompanyDomainResponse{PublicID: "dom_01"}, nil
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/companies/:id/domains/:domain_id", NewHandler(&fakeUpdateDomainService{}).Handle)
	body := []byte(`{"domain":"acme.com","kind":"alias","status":"active","redirect_to_primary":true}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/companies/cmp_01/domains/dom_01", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
