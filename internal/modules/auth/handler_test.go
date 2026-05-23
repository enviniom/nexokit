package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeAuthService struct {
	loginResult   *LoginResponse
	refreshResult *TokenPairResponse
	err           error
}

func (f fakeAuthService) Login(req LoginRequest) (*LoginResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.loginResult, nil
}

func (f fakeAuthService) Refresh(req RefreshRequest) (*TokenPairResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.refreshResult, nil
}

func (f fakeAuthService) Logout(req RefreshRequest) error { return f.err }

func authJSONRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("returns tokens and sanitized user", func(t *testing.T) {
		h := NewHandler(fakeAuthService{loginResult: &LoginResponse{AccessToken: "access", RefreshToken: "refresh", User: users.UserResponse{PublicID: "user1", Email: "user@example.com", RoleName: "admin"}}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", LoginRequest{Email: "user@example.com", Password: "Secret1!"})

		h.Login(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), "password_hash") {
			t.Fatalf("response leaked password fields: %s", w.Body.String())
		}
		var resp response.APIResponse[LoginResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.AccessToken != "access" || resp.Data.RefreshToken != "refresh" || resp.Data.User.Email != "user@example.com" {
			t.Fatalf("unexpected login response: %#v", resp.Data)
		}
	})

	t.Run("returns validation errors before service call", func(t *testing.T) {
		h := NewHandler(fakeAuthService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", LoginRequest{Email: "bad"})

		h.Login(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}
	})

	t.Run("returns generic unauthorized for invalid credentials", func(t *testing.T) {
		h := NewHandler(fakeAuthService{err: apperror.ErrUnauthorized})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", LoginRequest{Email: "user@example.com", Password: "Secret1!"})

		h.Login(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})
}

func TestHandler_RefreshLogoutAndMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("refresh returns rotated token pair", func(t *testing.T) {
		h := NewHandler(fakeAuthService{refreshResult: &TokenPairResponse{AccessToken: "new-access", RefreshToken: "new-refresh"}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/refresh", RefreshRequest{RefreshToken: "old-refresh"})

		h.Refresh(c)

		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "new-refresh") {
			t.Fatalf("expected rotated token response, got %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("logout returns success", func(t *testing.T) {
		h := NewHandler(fakeAuthService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/logout", RefreshRequest{RefreshToken: "refresh"})

		h.Logout(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("me returns context user with role and permission slugs", func(t *testing.T) {
		h := NewHandler(fakeAuthService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		authctx.SetGin(c, &authctx.User{PublicID: "user1", Email: "user@example.com", Name: "Alice", Role: "admin", RoleSlug: "admin", RoleID: 3, IsActive: true, Permissions: []string{"users.list", "auth.view"}})

		h.Me(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), "password_hash") {
			t.Fatalf("response leaked password fields: %s", w.Body.String())
		}
		var resp response.APIResponse[MeResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode me response: %v", err)
		}
		if resp.Data.RoleSlug != "admin" || len(resp.Data.Permissions) != 2 || resp.Data.Permissions[1] != "auth.view" {
			t.Fatalf("expected role slug and permissions in /me response, got %#v", resp.Data)
		}
	})

	t.Run("service errors map to unauthorized", func(t *testing.T) {
		h := NewHandler(fakeAuthService{err: errors.New("boom")})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/refresh", RefreshRequest{RefreshToken: "refresh"})

		h.Refresh(c)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}
