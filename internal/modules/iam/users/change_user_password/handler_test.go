package change_user_password

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct{ err error }

func (f fakeHandlerService) Change(_ tenant.TenantContext, _, _ string, _ core.ChangePasswordRequest) error {
	return f.err
}

func validChangePasswordBody() []byte {
	return []byte(`{"current_password":"oldpass123","new_password":"newpass123"}`)
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
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/iam/users/user-1/change-password", bytes.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "user-1"}}

			if tt.authUser != nil {
				authctx.SetGin(c, tt.authUser)
			}

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandlerHandleValidationErrorFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(fakeHandlerService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/iam/users/user-1/change-password", bytes.NewReader([]byte(`{"current_password":"","new_password":"short"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "user-1"}}

	h.Handle(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	errs, _ := payload["errors"].(map[string]any)
	if _, ok := errs["current_password"]; !ok {
		t.Fatalf("expected current_password field validation error")
	}
}
