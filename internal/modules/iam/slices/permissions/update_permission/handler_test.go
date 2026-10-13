package update_permission

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	item  *permission.PermissionResponse
	err   error
	calls *int
}

func (f fakeService) Update(_ string, _ permission.UpdatePermissionRequest) (*permission.PermissionResponse, error) {
	if f.calls != nil {
		*f.calls++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerTransportsWithoutActorPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name   string
		actor  *authctx.User
		status int
	}{
		{"no actor projection needed", nil, 200},
		{"actor projection does not authorize here", &authctx.User{Role: "root", RoleSlug: "root"}, 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.actor != nil {
				authctx.SetGin(c, tt.actor)
			}
			c.Request = httptest.NewRequest("PUT", "/permissions/custom", bytes.NewBufferString(`{"name":"Changed"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			NewHandler(fakeService{calls: &calls, item: &permission.PermissionResponse{}}).Handle(c)
			if w.Code != tt.status || calls != 1 {
				t.Fatalf("status=%d calls=%d, want %d and one", w.Code, calls, tt.status)
			}
		})
	}
}

func TestHandlerRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name   string
		body   string
		status int
	}{
		{name: "malformed JSON", body: `{"name":`, status: http.StatusBadRequest},
		{name: "wrong field type", body: `{"name":123}`, status: http.StatusBadRequest},
		{name: "invalid fields", body: `{}`, status: http.StatusUnprocessableEntity},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			h := NewHandler(fakeService{calls: &calls})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			authctx.SetGin(c, &authctx.User{IsRoot: true})
			c.Request = httptest.NewRequest(http.MethodPut, "/iam/permissions/perm-1", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Handle(c)
			if calls != 0 {
				t.Errorf("service calls = %d, want 0", calls)
			}
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			var payload map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload["message"] == nil || payload["message"] == "" {
				t.Errorf("missing rejection message: %#v", payload)
			}
			if payload["success"] != false {
				t.Errorf("success = %v, want false", payload["success"])
			}
			if tt.status == http.StatusUnprocessableEntity {
				fields, ok := payload["errors"].(map[string]any)
				if !ok || fields["name"] == nil {
					t.Errorf("expected field validation errors, got %v", payload["errors"])
				}
			}
		})
	}
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		service     Service
		statusCode  int
		wantMessage string
	}{
		{name: "returns success payload", service: fakeService{item: &permission.PermissionResponse{PublicID: "perm-1"}}, statusCode: http.StatusOK},
		{name: "maps not found to 404 with platform message", service: fakeService{err: core.ErrNotFound}, statusCode: http.StatusNotFound, wantMessage: "Recurso no encontrado"},
		{name: "maps immutable system permission to 403 with module message", service: fakeService{err: core.ErrSystemImmutable}, statusCode: http.StatusForbidden, wantMessage: "system resource is immutable"},
		{name: "maps conflict to 409 with platform message", service: fakeService{err: core.ErrConflict}, statusCode: http.StatusConflict, wantMessage: "El recurso ya existe"},
		{name: "maps unknown errors to 500", service: fakeService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	body := []byte(`{"name":"Manage","description":"desc","display_order":1}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			authctx.SetGin(c, &authctx.User{IsRoot: true})
			c.Params = gin.Params{{Key: "id", Value: "perm-1"}}
			c.Request = httptest.NewRequest(http.MethodPut, "/iam/permissions/perm-1", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			var payload map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if tt.wantMessage != "" {
				if got := payload["message"]; got != tt.wantMessage {
					t.Errorf("message = %q, want %q", got, tt.wantMessage)
				}
			}
		})
	}
}
