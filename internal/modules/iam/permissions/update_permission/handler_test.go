package update_permission

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	item *core.PermissionResponse
	err  error
}

func (f fakeService) Update(_ string, _ core.UpdatePermissionRequest) (*core.PermissionResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		statusCode int
	}{
		{name: "returns success payload", service: fakeService{item: &core.PermissionResponse{PublicID: "perm-1"}}, statusCode: http.StatusOK},
		{name: "maps not found to 404", service: fakeService{err: core.ErrNotFound}, statusCode: http.StatusNotFound},
		{name: "maps immutable system permission to 403", service: fakeService{err: core.ErrSystemImmutable}, statusCode: http.StatusForbidden},
		{name: "maps conflict to 409", service: fakeService{err: core.ErrConflict}, statusCode: http.StatusConflict},
		{name: "maps unknown errors to 500", service: fakeService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	body := []byte(`{"name":"Manage","description":"desc","display_order":1}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
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
		})
	}
}
