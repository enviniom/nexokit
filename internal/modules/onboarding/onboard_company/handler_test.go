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

func TestHandler_Onboard_ConflictErrors(t *testing.T) {
	for _, tt := range []struct {
		name         string
		serviceError error
		errorField   string
	}{
		{name: "duplicate company slug", serviceError: core.ErrDuplicateCompanySlug, errorField: "slug"},
		{name: "duplicate company domain", serviceError: core.ErrDuplicateCompanyDomain, errorField: "domain"},
		{name: "duplicate technical domain", serviceError: core.ErrDuplicateTechnicalDomain, errorField: "generate_technical_domain"},
		{name: "missing platform domain", serviceError: core.ErrMissingPlatformDomain, errorField: "generate_technical_domain"},
		{name: "duplicate admin email", serviceError: core.ErrDuplicateAdminEmail, errorField: "admin_email"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := setupOnboardingRouter(&authctx.User{RoleSlug: "root", IsRoot: true}, &fakeOnboardingService{err: tt.serviceError})
			w := httptest.NewRecorder()
			payload := core.OnboardCompanyRequest{Name: "Acme Corp", Slug: "acme", AdminName: "Jane Doe", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
			r.ServeHTTP(w, onboardJSONRequest(payload))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status 422 Unprocessable, got %d. Body: %s", w.Code, w.Body.String())
			}

			var resp response.APIResponse[any]
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			errs, ok := resp.Errors.(map[string]any)
			if !ok || errs[tt.errorField] == nil {
				t.Errorf("expected error field %q to be set, got: %#v", tt.errorField, resp.Errors)
			}
		})
	}
}
