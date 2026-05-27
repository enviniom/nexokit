package update_company

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/gin-gonic/gin"
)

type fakeUpdateService struct{ resp *core.CompanyResponse }

func (f *fakeUpdateService) Update(string, core.UpdateCompanyRequest) (*core.CompanyResponse, error) {
	return f.resp, nil
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/companies/:id", NewHandler(&fakeUpdateService{resp: &core.CompanyResponse{PublicID: "cmp_01"}}).Handle)
	w := httptest.NewRecorder()
	body := []byte(`{"name":"Acme","slug":"acme","status":"active"}`)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/companies/cmp_01", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
