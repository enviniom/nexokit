package view_company

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/gin-gonic/gin"
)

type fakeViewService struct{ resp *core.CompanyResponse }

func (f *fakeViewService) GetByPublicID(string) (*core.CompanyResponse, error) { return f.resp, nil }

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/companies/:id", NewHandler(&fakeViewService{resp: &core.CompanyResponse{PublicID: "cmp_01"}}).Handle)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/companies/cmp_01", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
