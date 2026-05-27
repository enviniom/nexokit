package create_company_domain

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/gin-gonic/gin"
)

type fakeCreateDomainService struct{}

func (f *fakeCreateDomainService) CreateDomain(string, core.CreateCompanyDomainRequest) (*core.CompanyDomainResponse, error) {
	return &core.CompanyDomainResponse{PublicID: "dom_01"}, nil
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/companies/:id/domains", NewHandler(&fakeCreateDomainService{}).Handle)
	body := []byte(`{"domain":"acme.com","kind":"primary","status":"active","redirect_to_primary":false}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/companies/cmp_01/domains", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
