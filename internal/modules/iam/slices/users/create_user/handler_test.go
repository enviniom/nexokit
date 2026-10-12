package create_user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	item  *iamuser.UserResponse
	err   error
	calls *int
}

func (f fakeHandlerService) Create(_ tenant.TenantContext, _ iamuser.CreateUserRequest) (*iamuser.UserResponse, error) {
	if f.calls != nil {
		(*f.calls)++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func validCreateBody() []byte {
	return []byte(`{"name":"Alice","email":"alice@example.com","password":"password123","role_id":2}`)
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(validCreateBody()))
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

	tests := []struct {
		name       string
		service    Service
		body       []byte
		statusCode int
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{
			name:       "success",
			service:    fakeHandlerService{item: &iamuser.UserResponse{PublicID: "user-1", Name: "Alice"}},
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
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				calls := 0
				service := tt.service.(fakeHandlerService)
				service.calls = &calls
				h := NewHandler(service)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/iam/users", bytes.NewReader(tt.body))
				c.Request.Header.Set("Content-Type", "application/json")

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}

				if tt.name == "invalid request body 400" || tt.name == "validation errors 422" {
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
						for _, field := range []string{"name", "email", "password", "role_id"} {
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
