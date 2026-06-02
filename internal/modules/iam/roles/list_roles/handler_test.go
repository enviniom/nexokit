package list_roles

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	items []core.RoleResponse
	total int64
	err   error
}

func (f fakeHandlerService) List(_ tenant.TenantContext, _ query.ListParams) ([]core.RoleResponse, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
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
			service:    fakeHandlerService{items: []core.RoleResponse{{PublicID: "role-1", Name: "Manager", Slug: "manager"}}, total: 1},
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
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/roles?page=1&per_page=20", nil)

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
