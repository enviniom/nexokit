package list_roles

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	items []iamrole.RoleResponse
	total int64
	err   error
}

func (f fakeHandlerService) List(_ tenant.TenantContext, _ query.ListParams) ([]iamrole.RoleResponse, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
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

func TestHandlerSafeListContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	company, actor := "company-public", uint(9)
	stamp := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	base := iamrole.RoleResponse{PublicID: "role-1", Name: "Manager", Slug: "manager", Description: "desc", IsSystem: true, CreatedAt: stamp, UpdatedAt: stamp}
	withOptional := base
	withOptional.CompanyID, withOptional.CreatedBy, withOptional.UpdatedBy = &company, &actor, &actor
	withOptional.Permissions = []string{"users.read"}
	emptyPermissions := base
	emptyPermissions.Permissions = []string{}
	for _, tt := range []struct {
		name  string
		items []iamrole.RoleResponse
		want  string
	}{
		{name: "nil list", want: `null`},
		{name: "empty list", items: []iamrole.RoleResponse{}, want: `[]`},
		{name: "optional fields omitted and nil permissions", items: []iamrole.RoleResponse{base}, want: `[{"public_id":"role-1","name":"Manager","slug":"manager","description":"desc","is_system":true,"permissions":null,"created_at":"2025-01-02T03:04:05Z","updated_at":"2025-01-02T03:04:05Z"}]`},
		{name: "empty permissions", items: []iamrole.RoleResponse{emptyPermissions}, want: `[{"public_id":"role-1","name":"Manager","slug":"manager","description":"desc","is_system":true,"permissions":[],"created_at":"2025-01-02T03:04:05Z","updated_at":"2025-01-02T03:04:05Z"}]`},
		{name: "public company and audit fields retained", items: []iamrole.RoleResponse{withOptional}, want: `[{"public_id":"role-1","company_id":"company-public","name":"Manager","slug":"manager","description":"desc","is_system":true,"permissions":["users.read"],"created_by":9,"updated_by":9,"created_at":"2025-01-02T03:04:05Z","updated_at":"2025-01-02T03:04:05Z"}]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/roles?page=1&per_page=20", nil)
			tenant.SetGin(c, tenant.NewRoot())
			NewHandler(fakeHandlerService{items: tt.items, total: int64(len(tt.items))}).Handle(c)
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
				t.Fatalf("safe data = %#v, want %#v (no private id/password/hash)", data, want)
			}
			meta, ok := payload["meta"].(map[string]any)
			if !ok {
				t.Fatalf("missing meta: %#v", payload)
			}
			pagination, ok := meta["pagination"].(map[string]any)
			if !ok || pagination["total"] != float64(len(tt.items)) {
				t.Fatalf("pagination = %#v", meta)
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
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{
			name:       "success paginated response",
			service:    fakeHandlerService{items: []iamrole.RoleResponse{{PublicID: "role-1", Name: "Manager", Slug: "manager"}}, total: 1},
			statusCode: http.StatusOK,
			assertBody: func(t *testing.T, payload map[string]any) {
				if payload["success"] != true {
					t.Fatalf("expected success true")
				}
				data, ok := payload["data"].([]any)
				if !ok || len(data) != 1 {
					t.Fatalf("expected one role in data")
				}
				meta, ok := payload["meta"].(map[string]any)
				if !ok {
					t.Fatalf("expected meta object")
				}
				pagination, ok := meta["pagination"].(map[string]any)
				if !ok {
					t.Fatalf("expected pagination meta")
				}
				if pagination["total"] != float64(1) {
					t.Fatalf("expected total 1, got %#v", pagination["total"])
				}
			},
		},
		{name: "service error 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				h := NewHandler(tt.service)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/iam/roles?page=1&per_page=20", nil)

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}

				var payload map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if tt.assertBody != nil {
					tt := tt.assertBody
					tt(t, payload)
				}
			})
		}
	}
}
