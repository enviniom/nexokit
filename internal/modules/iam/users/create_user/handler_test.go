package create_user

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
	item *core.UserResponse
	err  error
}

func (f fakeHandlerService) Create(_ tenant.TenantContext, _ core.CreateUserRequest) (*core.UserResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func validCreateBody() []byte {
	return []byte(`{"name":"Alice","email":"alice@example.com","password":"password123","role_id":2}`)
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		body       []byte
		statusCode int
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{
			name:       "success",
			service:    fakeHandlerService{item: &core.UserResponse{PublicID: "user-1", Name: "Alice"}},
			body:       validCreateBody(),
			statusCode: http.StatusCreated,
		},
		{
			name:       "duplicate email 422",
			service:    fakeHandlerService{err: core.ErrUserEmailAlreadyExists},
			body:       validCreateBody(),
			statusCode: http.StatusUnprocessableEntity,
			assertBody: func(t *testing.T, payload map[string]any) {
				errs, _ := payload["errors"].(map[string]any)
				if _, ok := errs["email"]; !ok {
					t.Fatalf("expected email field validation error")
				}
			},
		},
		{
			name:       "forbidden role assignment 403",
			service:    fakeHandlerService{err: core.ErrForbiddenRoleAssignment},
			body:       validCreateBody(),
			statusCode: http.StatusForbidden,
		},
		{
			name:       "invalid company scope 400",
			service:    fakeHandlerService{err: core.ErrInvalidCompanyScope},
			body:       validCreateBody(),
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "generic error 500",
			service:    fakeHandlerService{err: errors.New("db down")},
			body:       validCreateBody(),
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "invalid request body 400",
			service:    fakeHandlerService{},
			body:       []byte(`{not json`),
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "validation errors 422",
			service:    fakeHandlerService{},
			body:       []byte(`{"name":"","email":"","password":"","role_id":0}`),
			statusCode: http.StatusUnprocessableEntity,
			assertBody: func(t *testing.T, payload map[string]any) {
				errs, _ := payload["errors"].(map[string]any)
				if _, ok := errs["name"]; !ok {
					t.Fatalf("expected name field validation error")
				}
				if _, ok := errs["email"]; !ok {
					t.Fatalf("expected email field validation error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/iam/users", bytes.NewReader(tt.body))
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
