package onboard_company

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeOnboardingService struct {
	res *core.OnboardCompanyResponse
	err error
}

func (f *fakeOnboardingService) Onboard(ctx context.Context, req core.OnboardCompanyRequest) (*core.OnboardCompanyResponse, error) {
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
	svc := &fakeOnboardingService{res: &core.OnboardCompanyResponse{CompanyPublicID: "comp_123", CompanySlug: "acme", AdminPublicID: "usr_admin", AdminEmail: "jane@acme.com"}}
	r := setupOnboardingRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, svc)

	w := httptest.NewRecorder()
	payload := core.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
	r.ServeHTTP(w, onboardJSONRequest(payload))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.APIResponse[core.OnboardCompanyResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Data.CompanySlug != "acme" || resp.Data.AdminEmail != "jane@acme.com" {
		t.Errorf("unexpected response data: %#v", resp.Data)
	}
}

func TestHandler_Onboard_NonRootForbidden(t *testing.T) {
	for _, role := range []string{"admin", "user"} {
		t.Run(role, func(t *testing.T) {
			r := setupOnboardingRouter(&authctx.User{RoleSlug: role}, &fakeOnboardingService{})
			w := httptest.NewRecorder()
			payload := core.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
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
			r := setupOnboardingRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, &fakeOnboardingService{err: tt.err})
			w := httptest.NewRecorder()
			payload := core.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
			r.ServeHTTP(w, onboardJSONRequest(payload))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status 422 Unprocessable Entity, got %d. Body: %s", w.Code, w.Body.String())
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
