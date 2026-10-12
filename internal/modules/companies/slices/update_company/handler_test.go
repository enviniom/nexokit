package update_company

import (
	"bytes"
	"encoding/json"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeUpdateService struct {
	resp  *company.CompanyResponse
	err   error
	calls int
}

func (f *fakeUpdateService) Update(string, company.UpdateCompanyRequest) (*company.CompanyResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}

func TestHandler_RejectsInvalidInputBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct{ name, body, field string }{
		{"missing name", `{"slug":"acme","status":"active"}`, "name"},
		{"short slug", `{"name":"Acme","slug":"a","status":"active"}`, "slug"},
		{"invalid status", `{"name":"Acme","slug":"acme","status":"other"}`, "status"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeUpdateService{}
			r := gin.New()
			r.PUT("/companies/:id", NewHandler(svc).Handle)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/companies/cmp_01", bytes.NewBufferString(tt.body)))
			var resp response.ValidationErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if w.Code != http.StatusUnprocessableEntity || svc.calls != 0 || len(resp.Errors[tt.field]) == 0 {
				t.Fatalf("status=%d service calls=%d errors=%v", w.Code, svc.calls, resp.Errors)
			}
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/companies/:id", NewHandler(&fakeUpdateService{resp: &company.CompanyResponse{PublicID: "cmp_01"}}).Handle)
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
