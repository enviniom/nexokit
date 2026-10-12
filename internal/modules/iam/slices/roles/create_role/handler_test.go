package create_role

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	item  *iamrole.RoleResponse
	err   error
	calls *int
}

func (f fakeHandlerService) Create(_ tenant.TenantContext, _ iamrole.CreateRoleRequest) (*iamrole.RoleResponse, error) {
	if f.calls != nil {
		(*f.calls)++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"Manager","slug":"manager"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	// A nil service must remain unreachable when context is missing.
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("service invoked without tenant context: %v", p)
		}
	}()
	NewHandler(nil).Handle(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"name":"Manager","slug":"manager","description":"desc"}`)

	tests := []struct {
		name       string
		service    Service
		body       []byte
		statusCode int
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{name: "bad json", service: fakeHandlerService{}, body: []byte(`{`), statusCode: http.StatusBadRequest},
		{name: "request validation", service: fakeHandlerService{}, body: []byte(`{"name":"A","slug":"not valid"}`), statusCode: http.StatusUnprocessableEntity},
		{name: "success", service: fakeHandlerService{item: &iamrole.RoleResponse{PublicID: "role-1"}}, statusCode: http.StatusCreated},
		{name: "duplicate name 422", service: fakeHandlerService{err: core.ErrRoleNameAlreadyExists}, statusCode: http.StatusUnprocessableEntity, assertBody: func(t *testing.T, payload map[string]any) {
			errs, _ := payload["errors"].(map[string]any)
			if _, ok := errs["name"]; !ok {
				t.Fatalf("expected name field validation error")
			}
		}},
		{name: "duplicate slug 422", service: fakeHandlerService{err: core.ErrRoleSlugAlreadyExists}, statusCode: http.StatusUnprocessableEntity, assertBody: func(t *testing.T, payload map[string]any) {
			errs, _ := payload["errors"].(map[string]any)
			if _, ok := errs["slug"]; !ok {
				t.Fatalf("expected slug field validation error")
			}
		}},
		{name: "reserved identity 422", service: fakeHandlerService{err: core.ErrReservedRoleIdentity}, statusCode: http.StatusUnprocessableEntity, assertBody: func(t *testing.T, payload map[string]any) {
			errs, _ := payload["errors"].(map[string]any)
			if _, ok := errs["name"]; !ok {
				t.Fatalf("expected name field validation error")
			}
			if _, ok := errs["slug"]; !ok {
				t.Fatalf("expected slug field validation error")
			}
		}},
		{name: "generic error 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				calls := 0
				service := tt.service.(fakeHandlerService)
				service.calls = &calls
				h := NewHandler(service)
				requestBody := body
				if tt.body != nil {
					requestBody = tt.body
				}
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/iam/roles", bytes.NewReader(requestBody))
				c.Request.Header.Set("Content-Type", "application/json")

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}

				var payload map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if tt.body != nil {
					if calls != 0 {
						t.Fatalf("service calls = %d, want 0", calls)
					}
					if payload["success"] != false || payload["message"] == "" || payload["message"] == nil {
						t.Fatalf("invalid rejection envelope: %#v", payload)
					}
					if tt.statusCode == 422 {
						fields, ok := payload["errors"].(map[string]any)
						if !ok {
							t.Fatalf("missing field errors: %#v", payload)
						}
						for _, field := range []string{"name", "slug"} {
							if _, ok := fields[field]; !ok {
								t.Errorf("missing %s error", field)
							}
						}
					}
				}
				if tt.assertBody != nil {
					tt.assertBody(t, payload)
				}
			})
		}
	}
}
