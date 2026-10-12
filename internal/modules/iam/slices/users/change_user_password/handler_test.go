package change_user_password

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	err   error
	calls *int
}

func (f fakeHandlerService) Change(_ tenant.TenantContext, _, _ string, _ iamuser.ChangePasswordRequest) error {
	if f.calls != nil {
		(*f.calls)++
	}
	return f.err
}

func validChangePasswordBody() []byte {
	return []byte(`{"current_password":"oldpass123","new_password":"newpass123"}`)
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(validChangePasswordBody()))
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
		authUser   *authctx.User
		statusCode int
	}{
		{
			name:       "success",
			service:    fakeHandlerService{},
			body:       validChangePasswordBody(),
			statusCode: http.StatusOK,
		},
		{
			name:       "not found 404",
			service:    fakeHandlerService{err: core.ErrNotFound},
			body:       validChangePasswordBody(),
			statusCode: http.StatusNotFound,
		},
		{
			name:       "forbidden 403",
			service:    fakeHandlerService{err: core.ErrForbidden},
			body:       validChangePasswordBody(),
			statusCode: http.StatusForbidden,
		},
		{
			name:       "unauthorized 401",
			service:    fakeHandlerService{err: core.ErrUnauthorized},
			body:       validChangePasswordBody(),
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "generic error 500",
			service:    fakeHandlerService{err: errors.New("db down")},
			body:       validChangePasswordBody(),
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
			body:       []byte(`{"current_password":"","new_password":""}`),
			statusCode: http.StatusUnprocessableEntity,
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
				c.Request = httptest.NewRequest(http.MethodPut, "/iam/users/user-1/change-password", bytes.NewReader(tt.body))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Params = gin.Params{{Key: "id", Value: "user-1"}}

				if tt.authUser != nil {
					authctx.SetGin(c, tt.authUser)
				}

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
						for _, field := range []string{"current_password", "new_password"} {
							if _, ok := fields[field]; !ok {
								t.Errorf("missing %s error", field)
							}
						}
					}
				}
			})
		}
	}
}

func TestHandlerHandleValidationErrorFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	calls := 0
	h := NewHandler(fakeHandlerService{calls: &calls})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/iam/users/user-1/change-password", bytes.NewReader([]byte(`{"current_password":"","new_password":"short"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "user-1"}}

	h.Handle(c)

	if calls != 0 {
		t.Fatalf("service calls = %d, want 0", calls)
	}
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	errs, _ := payload["errors"].(map[string]any)
	for _, field := range []string{"current_password", "new_password"} {
		if _, ok := errs[field]; !ok {
			t.Errorf("missing %s validation error", field)
		}
	}
	if payload["success"] != false {
		t.Fatalf("expected rejection envelope: %#v", payload)
	}
}
