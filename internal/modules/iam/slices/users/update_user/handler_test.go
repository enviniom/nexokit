package update_user

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

func (f fakeHandlerService) Update(_ tenant.TenantContext, _, _ string, _ iamuser.UpdateUserRequest) (*iamuser.UserResponse, error) {
	if f.calls != nil {
		(*f.calls)++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func validUpdateBody() []byte {
	return []byte(`{"name":"Alice","email":"alice@example.com"}`)
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(`{"name":"Alice","email":"alice@example.com"}`))
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
			body:       validUpdateBody(),
			statusCode: http.StatusOK,
		},
		{
			name:       "duplicate email 422",
			service:    fakeHandlerService{err: core.ErrUserEmailAlreadyExists},
			body:       validUpdateBody(),
			statusCode: http.StatusUnprocessableEntity,
			assertBody: func(t *testing.T, payload map[string]any) {
				errs, _ := payload["errors"].(map[string]any)
				if _, ok := errs["email"]; !ok {
					t.Fatalf("expected email field validation error")
				}
			},
		},
		{
			name:       "not found 404",
			service:    fakeHandlerService{err: core.ErrNotFound},
			body:       validUpdateBody(),
			statusCode: http.StatusNotFound,
		},
		{
			name:       "forbidden 403",
			service:    fakeHandlerService{err: core.ErrForbiddenRoleAssignment},
			body:       validUpdateBody(),
			statusCode: http.StatusForbidden,
		},
		{
			name:       "invalid company scope 400",
			service:    fakeHandlerService{err: core.ErrInvalidCompanyScope},
			body:       validUpdateBody(),
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "generic error 500",
			service:    fakeHandlerService{err: errors.New("db down")},
			body:       validUpdateBody(),
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
			body:       []byte(`{"name":"","email":""}`),
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
				c.Request = httptest.NewRequest(http.MethodPut, "/iam/users/user-1", bytes.NewReader(tt.body))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Params = gin.Params{{Key: "id", Value: "user-1"}}

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
						for _, field := range []string{"name", "email"} {
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
