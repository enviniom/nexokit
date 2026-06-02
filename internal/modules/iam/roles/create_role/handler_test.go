package create_role

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	item *core.RoleResponse
	err  error
}

func (f fakeHandlerService) Create(_ tenant.TenantContext, _ core.CreateRoleRequest) (*core.RoleResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"name":"Manager","slug":"manager","description":"desc"}`)

	tests := []struct {
		name       string
		service    Service
		statusCode int
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{name: "success", service: fakeHandlerService{item: &core.RoleResponse{PublicID: "role-1"}}, statusCode: http.StatusCreated},
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
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/iam/roles", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			var payload map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if tt.assertBody != nil {
				tt.assertBody(t, payload)
			}
		})
	}
}
