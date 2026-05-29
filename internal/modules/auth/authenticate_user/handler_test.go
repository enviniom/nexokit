package authenticate_user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	result *core.LoginResponse
	err    error
}

func (f fakeService) Login(req core.LoginRequest) (*core.LoginResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func authJSONRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns tokens and sanitized user", func(t *testing.T) {
		h := NewHandler(fakeService{result: &core.LoginResponse{AccessToken: "access", RefreshToken: "refresh", User: core.AuthUserResponse{PublicID: "user1", Email: "user@example.com", RoleName: "admin"}}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", core.LoginRequest{Email: "user@example.com", Password: "Secret1!"})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), "password_hash") {
			t.Fatalf("response leaked password fields: %s", w.Body.String())
		}
		var resp response.APIResponse[core.LoginResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.AccessToken != "access" || resp.Data.RefreshToken != "refresh" || resp.Data.User.Email != "user@example.com" {
			t.Fatalf("unexpected login response: %#v", resp.Data)
		}
	})

	t.Run("returns validation errors before service call", func(t *testing.T) {
		h := NewHandler(fakeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", core.LoginRequest{Email: "bad"})

		h.Handle(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}
	})

	t.Run("returns generic unauthorized for invalid credentials", func(t *testing.T) {
		h := NewHandler(fakeService{err: apperror.ErrUnauthorized})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", core.LoginRequest{Email: "user@example.com", Password: "Secret1!"})

		h.Handle(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})
}
