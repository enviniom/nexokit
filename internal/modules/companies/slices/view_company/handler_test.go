package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeViewService struct{ resp *company.CompanyResponse }

func (f *fakeViewService) GetByPublicID(string) (*company.CompanyResponse, error) { return f.resp, nil }

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/companies/:id", NewHandler(&fakeViewService{resp: &company.CompanyResponse{PublicID: "cmp_01"}}).Handle)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/companies/cmp_01", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
