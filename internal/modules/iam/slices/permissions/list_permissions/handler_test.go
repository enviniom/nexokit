package list_permissions

import (
	"encoding/json"
	"errors"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	groups []permission.PermissionGroupResponse
	err    error
}

func (f fakeService) ListGrouped() ([]permission.PermissionGroupResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.groups, nil
}

func TestHandlerSafeGroupedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	actor := uint(9)
	stamp := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	base := permission.PermissionResponse{PublicID: "perm-1", Slug: "users.read", Name: "Read users", Module: "users", Action: "read", Description: "desc", IsSystem: true, DisplayOrder: 3, CreatedAt: stamp, UpdatedAt: stamp}
	withAudit := base
	withAudit.CreatedBy, withAudit.UpdatedBy = &actor, &actor
	for _, tt := range []struct {
		name   string
		groups []permission.PermissionGroupResponse
		want   string
	}{
		{name: "nil groups", want: `null`},
		{name: "empty groups", groups: []permission.PermissionGroupResponse{}, want: `[]`},
		{name: "nil permissions", groups: []permission.PermissionGroupResponse{{Module: "users"}}, want: `[{"module":"users","permissions":null}]`},
		{name: "empty permissions", groups: []permission.PermissionGroupResponse{{Module: "users", Permissions: []permission.PermissionResponse{}}}, want: `[{"module":"users","permissions":[]}]`},
		{name: "audit fields omitted", groups: []permission.PermissionGroupResponse{{Module: "users", Permissions: []permission.PermissionResponse{base}}}, want: `[{"module":"users","permissions":[{"public_id":"perm-1","slug":"users.read","name":"Read users","module":"users","action":"read","description":"desc","is_system":true,"display_order":3,"created_at":"2025-01-02T03:04:05Z","updated_at":"2025-01-02T03:04:05Z"}]}]`},
		{name: "audit fields retained", groups: []permission.PermissionGroupResponse{{Module: "users", Permissions: []permission.PermissionResponse{withAudit}}}, want: `[{"module":"users","permissions":[{"public_id":"perm-1","slug":"users.read","name":"Read users","module":"users","action":"read","description":"desc","is_system":true,"display_order":3,"created_by":9,"updated_by":9,"created_at":"2025-01-02T03:04:05Z","updated_at":"2025-01-02T03:04:05Z"}]}]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/permissions", nil)
			NewHandler(fakeService{groups: tt.groups}).Handle(c)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			var payload map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload["success"] != true {
				t.Fatalf("success envelope: %#v", payload)
			}
			var want any
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			data, exists := payload["data"]
			if !exists || !reflect.DeepEqual(data, want) {
				t.Fatalf("safe grouped data = %#v, want %#v (no private id/password/hash)", data, want)
			}
		})
	}
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
			service: fakeService{groups: []permission.PermissionGroupResponse{
				{Module: "roles", Permissions: []permission.PermissionResponse{{PublicID: "perm-1", Slug: "roles.list"}}},
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
