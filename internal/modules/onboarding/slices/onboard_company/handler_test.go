package onboard_company

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeOnboardingService struct {
	res   *companyentity.OnboardCompanyResponse
	err   error
	calls int
	req   companyentity.OnboardCompanyRequest
}

func (f *fakeOnboardingService) Onboard(ctx context.Context, req companyentity.OnboardCompanyRequest) (*companyentity.OnboardCompanyResponse, error) {
	f.calls++
	f.req = req
	if f.err != nil {
		return nil, f.err
	}
	return f.res, nil
}

func setupOnboardingRouter(user *authctx.User, svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if user != nil {
			authctx.SetGin(c, user)
		}
		c.Next()
	})

	requireRoleMW := func(role string) gin.HandlerFunc {
		return func(c *gin.Context) {
			u, ok := authctx.FromGin(c)
			if !ok || u == nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if u.RoleSlug != role {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
		}
	}

	v1 := r.Group("/api/v1")
	g := v1.Group("/onboarding")
	g.POST("/companies", requireRoleMW("root"), NewHandler(svc).Handle)
	return r
}

func onboardJSONRequest(body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/companies", &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_Onboard_RootSuccess(t *testing.T) {
	svc := &fakeOnboardingService{res: &companyentity.OnboardCompanyResponse{CompanyPublicID: "comp_123", CompanySlug: "acme", AdminPublicID: "usr_admin", AdminEmail: "jane@acme.com"}}
	r := setupOnboardingRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

	w := httptest.NewRecorder()
	payload := companyentity.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
	r.ServeHTTP(w, onboardJSONRequest(payload))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.APIResponse[companyentity.OnboardCompanyResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Message != messages.MsgCreated || resp.Data != *svc.res {
		t.Errorf("unexpected response data: %#v", resp.Data)
	}
	var envelope struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"company_public_id": "comp_123", "company_slug": "acme", "admin_public_id": "usr_admin", "admin_email": "jane@acme.com"}
	if !reflect.DeepEqual(envelope.Data, want) {
		t.Fatalf("public response fields=%v want=%v", envelope.Data, want)
	}
}

func TestHandler_Onboard_RequestValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*companyentity.OnboardCompanyRequest)
		fields []string
	}{
		{"required fields missing", func(r *companyentity.OnboardCompanyRequest) { *r = companyentity.OnboardCompanyRequest{} }, []string{"name", "slug", "admin_name", "admin_email", "admin_password"}},
		{"company name too short", func(r *companyentity.OnboardCompanyRequest) { r.Name = "A" }, []string{"name"}},
		{"slug too short", func(r *companyentity.OnboardCompanyRequest) { r.Slug = "a" }, []string{"slug"}},
		{"admin name too short", func(r *companyentity.OnboardCompanyRequest) { r.AdminName = "J" }, []string{"admin_name"}},
		{"invalid admin email", func(r *companyentity.OnboardCompanyRequest) { r.AdminEmail = "not-an-email" }, []string{"admin_email"}},
		{"password too short", func(r *companyentity.OnboardCompanyRequest) { r.AdminPassword = "short" }, []string{"admin_password"}},
		{"multiple invalid fields", func(r *companyentity.OnboardCompanyRequest) {
			r.Name = ""
			r.Slug = "a"
			r.AdminName = ""
			r.AdminEmail = "bad-email"
			r.AdminPassword = "tiny"
		}, []string{"name", "slug", "admin_name", "admin_email", "admin_password"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeOnboardingService{}
			r := setupOnboardingRouter(&authctx.User{RoleSlug: "root"}, svc)
			payload := baseRequest()
			tt.mutate(&payload)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, onboardJSONRequest(payload))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			var resp response.ValidationErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Success || resp.Message != messages.MsgValidationError || len(resp.Errors) != len(tt.fields) {
				t.Fatalf("unexpected validation response: %#v", resp)
			}
			for _, field := range tt.fields {
				if len(resp.Errors[field]) == 0 {
					t.Errorf("missing validation message for %s: %v", field, resp.Errors)
				}
			}
			if svc.calls != 0 {
				t.Fatal("invalid request reached provisioning service")
			}
		})
	}
}

func TestHandler_Onboard_PreservesRequestForService(t *testing.T) {
	domain := " ACME.TEST "
	payload := baseRequest()
	payload.Name, payload.Slug, payload.AdminName, payload.Domain = " Acme ", " ACME ", " Jane ", &domain
	svc := &fakeOnboardingService{res: &companyentity.OnboardCompanyResponse{}}
	r := setupOnboardingRouter(&authctx.User{RoleSlug: "root"}, svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, onboardJSONRequest(payload))
	if w.Code != http.StatusCreated || svc.calls != 1 || !reflect.DeepEqual(svc.req, payload) {
		t.Fatalf("status=%d calls=%d service request=%#v", w.Code, svc.calls, svc.req)
	}
}

func TestHandler_Onboard_MalformedJSON(t *testing.T) {
	svc := &fakeOnboardingService{}
	r := setupOnboardingRouter(&authctx.User{RoleSlug: "root"}, svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/companies", strings.NewReader(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest || svc.calls != 0 {
		t.Fatalf("status=%d calls=%d body=%s", w.Code, svc.calls, w.Body.String())
	}
}

func TestHandler_Onboard_NonRootForbidden(t *testing.T) {
	for _, role := range []string{"admin", "user"} {
		t.Run(role, func(t *testing.T) {
			r := setupOnboardingRouter(&authctx.User{RoleSlug: role}, &fakeOnboardingService{})
			w := httptest.NewRecorder()
			payload := companyentity.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
			r.ServeHTTP(w, onboardJSONRequest(payload))
			if w.Code != http.StatusForbidden {
				t.Fatalf("expected status 403 Forbidden for role %s, got %d", role, w.Code)
			}
		})
	}
}

func TestHandler_Onboard_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantField string
		wantMsg   string
	}{
		{name: "duplicate company slug", err: core.ErrDuplicateCompanySlug, wantField: "slug", wantMsg: "El recurso ya existe"},
		{name: "duplicate company domain", err: core.ErrDuplicateCompanyDomain, wantField: "domain", wantMsg: "El recurso ya existe"},
		{name: "duplicate technical domain", err: core.ErrDuplicateTechnicalDomain, wantField: "generate_technical_domain", wantMsg: "El recurso ya existe"},
		{name: "missing platform domain", err: core.ErrMissingPlatformDomain, wantField: "generate_technical_domain", wantMsg: "formato inválido"},
		{name: "duplicate admin email", err: core.ErrDuplicateAdminEmail, wantField: "admin_email", wantMsg: "El recurso ya existe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupOnboardingRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, &fakeOnboardingService{err: fmt.Errorf("onboarding failed: %w", tt.err)})
			w := httptest.NewRecorder()
			payload := companyentity.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
			r.ServeHTTP(w, onboardJSONRequest(payload))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status 422 Unprocessable Entity, got %d. Body: %s", w.Code, w.Body.String())
			}

			var resp response.ValidationErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if resp.Success || resp.Message != messages.MsgValidationError || resp.Data != nil || len(resp.Errors) != 1 {
				t.Errorf("unexpected validation response: %#v", resp)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
				t.Fatal(err)
			}
			if _, exists := fields["code"]; exists {
				t.Error("validation response unexpectedly emits a sentinel code")
			}
			fieldErrs, ok := resp.Errors[tt.wantField]
			if !ok {
				t.Fatalf("expected field %q in errors, got %v", tt.wantField, resp.Errors)
			}
			found := false
			for _, m := range fieldErrs {
				if m == tt.wantMsg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected message %q for field %q, got %v", tt.wantMsg, tt.wantField, fieldErrs)
			}
		})
	}
}
