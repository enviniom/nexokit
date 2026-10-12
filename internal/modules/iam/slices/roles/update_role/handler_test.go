package update_role

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

func (f fakeHandlerService) Update(_ tenant.TenantContext, _ string, _ iamrole.UpdateRoleRequest) (*iamrole.RoleResponse, error) {
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
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(`{"name":"Manager","slug":"manager"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("service invoked without context: %v", p)
		}
	}()
	NewHandler(nil).Handle(c)
	if w.Code != 400 {
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
		{name: "success", service: fakeHandlerService{item: &iamrole.RoleResponse{PublicID: "role-1"}}, body: body, statusCode: http.StatusOK},
		{name: "bad json", service: fakeHandlerService{}, body: []byte("{"), statusCode: http.StatusBadRequest},
		{name: "request validation", service: fakeHandlerService{}, body: []byte(`{"name":"A","slug":"not valid"}`), statusCode: http.StatusUnprocessableEntity},
		{name: "not found", service: fakeHandlerService{err: core.ErrNotFound}, body: body, statusCode: http.StatusNotFound},
		{name: "protected", service: fakeHandlerService{err: core.ErrRoleProtected}, body: body, statusCode: http.StatusForbidden},
		{name: "reserved identity", service: fakeHandlerService{err: core.ErrReservedRoleIdentity}, body: body, statusCode: http.StatusUnprocessableEntity, assertBody: func(t *testing.T, payload map[string]any) {
			errs, _ := payload["errors"].(map[string]any)
			if _, ok := errs["name"]; !ok {
				t.Fatalf("expected name field validation error")
			}
			if _, ok := errs["slug"]; !ok {
				t.Fatalf("expected slug field validation error")
			}
		}},
		{name: "duplicate name", service: fakeHandlerService{err: core.ErrRoleNameAlreadyExists}, body: body, statusCode: http.StatusUnprocessableEntity, assertBody: func(t *testing.T, payload map[string]any) {
			errs, _ := payload["errors"].(map[string]any)
			if _, ok := errs["name"]; !ok {
				t.Fatalf("expected name field validation error")
			}
		}},
		{name: "duplicate slug", service: fakeHandlerService{err: core.ErrRoleSlugAlreadyExists}, body: body, statusCode: http.StatusUnprocessableEntity, assertBody: func(t *testing.T, payload map[string]any) {
			errs, _ := payload["errors"].(map[string]any)
			if _, ok := errs["slug"]; !ok {
				t.Fatalf("expected slug field validation error")
			}
		}},
		{name: "generic error", service: fakeHandlerService{err: errors.New("db down")}, body: body, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				calls := 0
				service := tt.service.(fakeHandlerService)
				service.calls = &calls
				h := NewHandler(service)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{{Key: "id", Value: "role-1"}}
				c.Request = httptest.NewRequest(http.MethodPut, "/iam/roles/role-1", bytes.NewReader(tt.body))
				c.Request.Header.Set("Content-Type", "application/json")

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}

				if tt.name == "bad json" || tt.name == "request validation" {
					if calls != 0 {
						t.Fatalf("service calls = %d, want 0", calls)
					}
					var envelope map[string]any
					if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
						t.Fatal(err)
					}
					if envelope["success"] != false || envelope["message"] == "" || envelope["message"] == nil {
						t.Fatalf("invalid rejection envelope: %#v", envelope)
					}
					if tt.statusCode == 422 {
						fields, ok := envelope["errors"].(map[string]any)
						if !ok {
							t.Fatalf("missing field errors: %#v", envelope)
						}
						for _, field := range []string{"name", "slug"} {
							if _, ok := fields[field]; !ok {
								t.Errorf("missing %s error", field)
							}
						}
					}
				}
				if tt.assertBody != nil {
					var payload map[string]any
					if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					tt.assertBody(t, payload)
				}
			})
		}
	}
}
