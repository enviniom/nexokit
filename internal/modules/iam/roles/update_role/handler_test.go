package update_role

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

func (f fakeHandlerService) Update(_ tenant.TenantContext, _ string, _ core.UpdateRoleRequest) (*core.RoleResponse, error) {
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
		body       []byte
		statusCode int
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{name: "success", service: fakeHandlerService{item: &core.RoleResponse{PublicID: "role-1"}}, body: body, statusCode: http.StatusOK},
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
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "role-1"}}
			c.Request = httptest.NewRequest(http.MethodPut, "/iam/roles/role-1", bytes.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
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
