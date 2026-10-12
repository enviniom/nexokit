package create_company_domain

import (
	"bytes"
	"encoding/json"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeCreateDomainService struct {
	err   error
	calls int
}

func (f *fakeCreateDomainService) CreateDomain(string, company_domain.CreateCompanyDomainRequest) (*company_domain.CompanyDomainResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &company_domain.CompanyDomainResponse{PublicID: "dom_01"}, nil
}

func TestHandler_RejectsInvalidInputBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct{ name, body, field string }{
		{"empty normalized domain", `{"domain":"   ","kind":"alias","status":"active"}`, "domain"},
		{"primary redirect", `{"domain":"acme.com","kind":"primary","status":"active","redirect_to_primary":true}`, "redirect_to_primary"},
		{"invalid kind", `{"domain":"acme.com","kind":"other","status":"active"}`, "kind"},
		{"invalid status", `{"domain":"acme.com","kind":"alias","status":"other"}`, "status"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeCreateDomainService{}
			r := gin.New()
			r.POST("/companies/:id/domains", NewHandler(svc).Handle)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/companies/cmp_01/domains", bytes.NewBufferString(tt.body)))
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
	r.POST("/companies/:id/domains", NewHandler(&fakeCreateDomainService{}).Handle)
	body := []byte(`{"domain":"acme.com","kind":"primary","status":"active","redirect_to_primary":false}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/companies/cmp_01/domains", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
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
			r.POST("/companies/:id/domains", NewHandler(&fakeCreateDomainService{err: tt.err}).Handle)
			body := []byte(`{"domain":"acme.com","kind":"primary","status":"active","redirect_to_primary":false}`)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/companies/cmp_01/domains", bytes.NewReader(body)))

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
