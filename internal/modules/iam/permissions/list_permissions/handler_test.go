package list_permissions

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	groups []core.PermissionGroupResponse
	err    error
}

func (f fakeService) ListGrouped() ([]core.PermissionGroupResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.groups, nil
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		statusCode int
		assertBody func(t *testing.T, body map[string]any)
	}{
		{
			name: "returns grouped permissions on success",
			service: fakeService{groups: []core.PermissionGroupResponse{
				{Module: "roles", Permissions: []core.PermissionResponse{{PublicID: "perm-1", Slug: "roles.list"}}},
			}},
			statusCode: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["success"] != true {
					t.Fatalf("expected success=true, got %#v", body["success"])
				}
				data, ok := body["data"].([]any)
				if !ok || len(data) != 1 {
					t.Fatalf("expected one module group in data, got %#v", body["data"])
				}
				group, ok := data[0].(map[string]any)
				if !ok || group["module"] != "roles" {
					t.Fatalf("expected module 'roles', got %#v", data[0])
				}
			},
		},
		{
			name:       "maps app errors through standard error response",
			service:    fakeService{err: apperror.Wrap(apperror.ErrBadRequest, "invalid query")},
			statusCode: http.StatusBadRequest,
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["success"] != false {
					t.Fatalf("expected success=false, got %#v", body["success"])
				}
				if body["message"] != "invalid query" {
					t.Fatalf("expected message 'invalid query', got %#v", body["message"])
				}
			},
		},
		{
			name:       "falls back to 500 for generic errors",
			service:    fakeService{err: errors.New("database unavailable")},
			statusCode: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["success"] != false {
					t.Fatalf("expected success=false, got %#v", body["success"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/permissions", nil)

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			tt.assertBody(t, body)
		})
	}
}
