package update_company_domain

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

type fakeUpdateDomainService struct{ err error }

func (f *fakeUpdateDomainService) UpdateDomain(string, string, core.UpdateCompanyDomainRequest) (*core.CompanyDomainResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
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

func TestHandler_Handle_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantField string
	}{
		{name: "duplicate company domain", err: core.ErrDuplicateCompanyDomain, wantField: "domain"},
		{name: "active primary domain exists", err: core.ErrActivePrimaryDomainExists, wantField: "kind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.PUT("/companies/:id/domains/:domain_id", NewHandler(&fakeUpdateDomainService{err: tt.err}).Handle)
			body := []byte(`{"domain":"acme.com","kind":"alias","status":"active","redirect_to_primary":true}`)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/companies/cmp_01/domains/dom_01", bytes.NewReader(body)))

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
			fieldErrs, ok := resp.Errors[tt.wantField]
			if !ok {
				t.Fatalf("expected field %q in errors, got %v", tt.wantField, resp.Errors)
			}
			found := false
			for _, m := range fieldErrs {
				if m == messages.MsgConflict {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected message %q for field %q, got %v", messages.MsgConflict, tt.wantField, fieldErrs)
			}
		})
	}
}

func TestHandler_Handle_DomainDoesNotBelong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/companies/:id/domains/:domain_id", NewHandler(&fakeUpdateDomainService{err: core.ErrCompanyDomainDoesNotBelong}).Handle)
	body := []byte(`{"domain":"acme.com","kind":"alias","status":"active","redirect_to_primary":true}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/companies/cmp_01/domains/dom_01", bytes.NewReader(body)))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Success {
		t.Errorf("expected error response")
	}
	if resp.Message != messages.MsgNotFound {
		t.Errorf("expected message %q, got %q", messages.MsgNotFound, resp.Message)
	}
}
