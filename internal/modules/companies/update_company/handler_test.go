package update_company

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeUpdateService struct {
	resp *core.CompanyResponse
	err  error
}

func (f *fakeUpdateService) Update(string, core.UpdateCompanyRequest) (*core.CompanyResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
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

func TestHandler_Handle_DuplicateSlugValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/companies/:id", NewHandler(&fakeUpdateService{err: core.ErrDuplicateCompanySlug}).Handle)
	w := httptest.NewRecorder()
	body := []byte(`{"name":"Acme","slug":"acme","status":"active"}`)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/companies/cmp_01", bytes.NewReader(body)))

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.ValidationErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Success {
		t.Errorf("expected error response")
	}
	fieldErrs, ok := resp.Errors["slug"]
	if !ok {
		t.Fatalf("expected field %q in errors, got %v", "slug", resp.Errors)
	}
	found := false
	for _, m := range fieldErrs {
		if m == messages.MsgConflict {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected message %q for field %q, got %v", messages.MsgConflict, "slug", fieldErrs)
	}
}
