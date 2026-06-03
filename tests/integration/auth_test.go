package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth"
	authcore "github.com/enviniom/nexokit/internal/modules/auth/core"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/password"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/token"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/enviniom/nexokit/tests/helpers"
	"github.com/gin-gonic/gin"
)

func TestAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)
	db := helpers.NewSQLiteDB(t, &iamcore.IAMRole{}, &iamcore.IAMUser{}, &authcore.RefreshToken{})
	adminRole := helpers.SeedRole(t, db, iamcore.AdminRoleSlug)
	pw := password.Manager{}
	hash, err := pw.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	activeUser := iamcore.IAMUser{BaseModel: shared.BaseModel{PublicID: "auth-active"}, Name: "Active", Email: "active@example.com", PasswordHash: hash, RoleID: adminRole.ID, IsActive: true}
	inactiveUser := iamcore.IAMUser{BaseModel: shared.BaseModel{PublicID: "auth-inactive"}, Name: "Inactive", Email: "inactive@example.com", PasswordHash: hash, RoleID: adminRole.ID, IsActive: false}
	if err := db.Create(&activeUser).Error; err != nil {
		t.Fatalf("seed active user: %v", err)
	}
	if err := db.Create(&inactiveUser).Error; err != nil {
		t.Fatalf("seed inactive user: %v", err)
	}
	if err := db.Model(&iamcore.IAMUser{}).Where("id = ?", inactiveUser.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("mark inactive user: %v", err)
	}

	manager := token.NewManager("nexokit-test-secret", time.Hour)
	c := auth.NewContainer(db, pw, manager, 24*time.Hour)
	r := gin.New()
	v1 := r.Group("/api/v1")
	auth.Register(v1, c, func(c *gin.Context) {
		authctx.SetGin(c, &authctx.User{PublicID: activeUser.PublicID, Name: activeUser.Name, Email: activeUser.Email, Role: adminRole.Name, RoleSlug: adminRole.Slug, RoleID: adminRole.ID, CompanyID: activeUser.CompanyID, IsActive: true, Permissions: []string{"users.list", "roles.view"}})
		c.Next()
	}, func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() })

	t.Run("login success returns token pair", func(t *testing.T) {
		w := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp response.APIResponse[authcore.LoginResponse]
		mustDecode(t, w.Body.Bytes(), &resp)
		if resp.Data.AccessToken == "" || resp.Data.RefreshToken == "" {
			t.Fatalf("expected access and refresh tokens")
		}
	})

	t.Run("invalid credentials return 401", func(t *testing.T) {
		w := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "wrong-password"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("inactive user login returns 401", func(t *testing.T) {
		w := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "inactive@example.com", "password": "secret123"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("valid refresh rotates token", func(t *testing.T) {
		login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		var loginResp response.APIResponse[authcore.LoginResponse]
		mustDecode(t, login.Body.Bytes(), &loginResp)

		refresh := requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		if refresh.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", refresh.Code, refresh.Body.String())
		}
		var rotated response.APIResponse[authcore.TokenPairResponse]
		mustDecode(t, refresh.Body.Bytes(), &rotated)
		if rotated.Data.RefreshToken == "" || rotated.Data.RefreshToken == loginResp.Data.RefreshToken {
			t.Fatalf("expected rotated refresh token")
		}
	})

	t.Run("revoked refresh token returns 401", func(t *testing.T) {
		login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		var loginResp response.APIResponse[authcore.LoginResponse]
		mustDecode(t, login.Body.Bytes(), &loginResp)

		_ = requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		reuse := requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		if reuse.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for revoked token, got %d: %s", reuse.Code, reuse.Body.String())
		}
	})

	t.Run("logout revokes refresh token", func(t *testing.T) {
		login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		var loginResp response.APIResponse[authcore.LoginResponse]
		mustDecode(t, login.Body.Bytes(), &loginResp)

		logout := requestJSON(r, http.MethodPost, "/api/v1/auth/logout", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		if logout.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", logout.Code, logout.Body.String())
		}

		reuse := requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		if reuse.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 after logout revoke, got %d: %s", reuse.Code, reuse.Body.String())
		}
	})

	t.Run("me returns authenticated session", func(t *testing.T) {
		me := requestJSON(r, http.MethodGet, "/api/v1/auth/me", nil)
		if me.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", me.Code, me.Body.String())
		}
		var resp response.APIResponse[authcore.MeResponse]
		mustDecode(t, me.Body.Bytes(), &resp)
		if resp.Data.PublicID != activeUser.PublicID || len(resp.Data.Permissions) != 2 {
			t.Fatalf("unexpected /me payload: %#v", resp.Data)
		}
	})
}

func requestJSON(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func mustDecode(t *testing.T, payload []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}
