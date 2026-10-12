package integration_test

import (
	"bytes"
	"encoding/json"
	userentity "github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"github.com/enviniom/nexokit/internal/modules/iam/core/refresh_token"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/iam"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
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
	db := helpers.NewSQLiteDB(t, &iamrole.IAMRole{}, &permission.IAMPermission{}, &iamuser.IAMUser{}, &refresh_token.RefreshToken{})
	adminRole := helpers.SeedRole(t, db, iamcore.AdminRoleSlug)
	permissions := []permission.IAMPermission{
		{BaseModel: shared.BaseModel{PublicID: "auth-permission-users-list"}, Slug: "users.list", Name: "List users", Module: "users", Action: "list"},
		{BaseModel: shared.BaseModel{PublicID: "auth-permission-roles-view"}, Slug: "roles.view", Name: "View roles", Module: "roles", Action: "view"},
	}
	if err := db.Model(&adminRole).Association("Permissions").Append(&permissions); err != nil {
		t.Fatalf("seed role permissions: %v", err)
	}
	pw := password.Manager{}
	hash, err := pw.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	activeUser := iamuser.IAMUser{BaseModel: shared.BaseModel{PublicID: "auth-active"}, Name: "Active", Email: "active@example.com", PasswordHash: hash, RoleID: adminRole.ID, IsActive: true}
	inactiveUser := iamuser.IAMUser{BaseModel: shared.BaseModel{PublicID: "auth-inactive"}, Name: "Inactive", Email: "inactive@example.com", PasswordHash: hash, RoleID: adminRole.ID, IsActive: false}
	if err := db.Create(&activeUser).Error; err != nil {
		t.Fatalf("seed active user: %v", err)
	}
	if err := db.Create(&inactiveUser).Error; err != nil {
		t.Fatalf("seed inactive user: %v", err)
	}
	if err := db.Model(&iamuser.IAMUser{}).Where("id = ?", inactiveUser.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("mark inactive user: %v", err)
	}

	manager := token.NewManager("nexokit-test-secret", time.Hour)
	iamContainer := iam.NewContainer(db, nil, pw, manager, 24*time.Hour)
	r := gin.New()
	v1 := r.Group("/api/v1")
	pass := func(c *gin.Context) { c.Next() }
	iam.Register(v1, v1.Group(""), v1.Group(""), iamContainer, func(string) gin.HandlerFunc { return pass }, middleware.Auth(manager, integrationAuthUserLookup{resolver: iamContainer.AuthUserResolver}), pass, pass)

	t.Run("login success returns token pair", func(t *testing.T) {
		w := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp response.APIResponse[userentity.LoginResponse]
		mustDecode(t, w.Body.Bytes(), &resp)
		if resp.Data.AccessToken == "" || resp.Data.RefreshToken == "" {
			t.Fatalf("expected access and refresh tokens")
		}
		var envelope struct {
			Data map[string]json.RawMessage `json:"data"`
		}
		mustDecode(t, w.Body.Bytes(), &envelope)
		if len(envelope.Data) != 3 {
			t.Fatalf("unexpected login fields: %s", w.Body.String())
		}
		assertAuthUserJSON(t, envelope.Data["user"], activeUser, adminRole.Name, false)
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
		var loginResp response.APIResponse[userentity.LoginResponse]
		mustDecode(t, login.Body.Bytes(), &loginResp)

		refresh := requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		if refresh.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", refresh.Code, refresh.Body.String())
		}
		var rotated response.APIResponse[refresh_token.TokenPairResponse]
		mustDecode(t, refresh.Body.Bytes(), &rotated)
		if rotated.Data.RefreshToken == "" || rotated.Data.RefreshToken == loginResp.Data.RefreshToken {
			t.Fatalf("expected rotated refresh token")
		}
	})

	for _, tc := range []struct {
		name     string
		publicID string
		expire   bool
	}{
		{name: "expired refresh returns 401 without replacement", publicID: "auth-refresh-expired", expire: true},
		{name: "refresh after user deactivation returns 401 without replacement", publicID: "auth-refresh-deactivated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user := iamuser.IAMUser{BaseModel: shared.BaseModel{PublicID: tc.publicID}, Name: "Refresh rejection", Email: tc.publicID + "@example.com", PasswordHash: hash, RoleID: adminRole.ID, IsActive: true}
			if err := db.Create(&user).Error; err != nil {
				t.Fatalf("seed refresh user: %v", err)
			}
			login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": user.Email, "password": "secret123"})
			if login.Code != http.StatusOK {
				t.Fatalf("expected login 200, got %d: %s", login.Code, login.Body.String())
			}
			var loginResp response.APIResponse[userentity.LoginResponse]
			mustDecode(t, login.Body.Bytes(), &loginResp)
			if loginResp.Data.RefreshToken == "" {
				t.Fatal("expected login refresh token")
			}

			if tc.expire {
				// Expire the persisted session deterministically, without waiting for TTL.
				if err := db.Model(&refresh_token.RefreshToken{}).Where("user_id = ?", user.ID).Update("expires_at", time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)).Error; err != nil {
					t.Fatalf("expire refresh session: %v", err)
				}
			} else {
				if err := db.Model(&user).Update("is_active", false).Error; err != nil {
					t.Fatalf("deactivate user after login: %v", err)
				}
			}
			var before []refresh_token.RefreshToken
			if err := db.Where("user_id = ?", user.ID).Order("id").Find(&before).Error; err != nil {
				t.Fatalf("load sessions before refresh: %v", err)
			}
			if len(before) != 1 || before[0].TokenHash == "" || before[0].RevokedAt != nil || before[0].ReplacedByHash != nil {
				t.Fatalf("expected one unrotated login session, got %#v", before)
			}

			refresh := requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
			if refresh.Code != http.StatusUnauthorized {
				t.Fatalf("expected refresh 401, got %d: %s", refresh.Code, refresh.Body.String())
			}
			var after []refresh_token.RefreshToken
			if err := db.Where("user_id = ?", user.ID).Order("id").Find(&after).Error; err != nil {
				t.Fatalf("load sessions after refresh: %v", err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("rejected refresh changed session count, hash, or state: before=%#v after=%#v", before, after)
			}
		})
	}

	t.Run("revoked refresh token returns 401", func(t *testing.T) {
		login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		var loginResp response.APIResponse[userentity.LoginResponse]
		mustDecode(t, login.Body.Bytes(), &loginResp)

		_ = requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		reuse := requestJSON(r, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": loginResp.Data.RefreshToken})
		if reuse.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for revoked token, got %d: %s", reuse.Code, reuse.Body.String())
		}
	})

	t.Run("logout revokes refresh token", func(t *testing.T) {
		login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		var loginResp response.APIResponse[userentity.LoginResponse]
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

	for _, tc := range []struct {
		name          string
		authorization string
	}{
		{name: "me rejects missing bearer"},
		{name: "me rejects invalid bearer", authorization: "Bearer invalid-token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
			if tc.authorization != "" {
				req.Header.Set("Authorization", tc.authorization)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
			}
		})
	}

	t.Run("me returns authenticated session with valid bearer", func(t *testing.T) {
		login := requestJSON(r, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "active@example.com", "password": "secret123"})
		if login.Code != http.StatusOK {
			t.Fatalf("expected login 200, got %d: %s", login.Code, login.Body.String())
		}
		var loginResp response.APIResponse[userentity.LoginResponse]
		mustDecode(t, login.Body.Bytes(), &loginResp)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
		me := httptest.NewRecorder()
		r.ServeHTTP(me, req)
		if me.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", me.Code, me.Body.String())
		}
		var resp response.APIResponse[userentity.MeResponse]
		mustDecode(t, me.Body.Bytes(), &resp)
		if resp.Data.PublicID != activeUser.PublicID || len(resp.Data.Permissions) != 2 {
			t.Fatalf("unexpected /me payload: %#v", resp.Data)
		}
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		mustDecode(t, me.Body.Bytes(), &envelope)
		assertAuthUserJSON(t, envelope.Data, activeUser, adminRole.Name, true)
		gotPermissions := make(map[string]bool)
		for _, slug := range resp.Data.Permissions {
			gotPermissions[slug] = true
		}
		if !gotPermissions["users.list"] || !gotPermissions["roles.view"] {
			t.Fatalf("unexpected /me permissions: %v", resp.Data.Permissions)
		}
	})
}

func assertAuthUserJSON(t *testing.T, payload []byte, model iamuser.IAMUser, roleName string, me bool) {
	t.Helper()
	var fields map[string]json.RawMessage
	mustDecode(t, payload, &fields)
	expected := map[string]any{
		"public_id": model.PublicID, "name": model.Name, "email": model.Email,
		"is_active": true, "role_id": model.RoleID, "role_name": roleName,
		"created_at": model.CreatedAt, "updated_at": model.UpdatedAt,
	}
	if me {
		expected["created_at"] = time.Time{}
		expected["updated_at"] = time.Time{}
		expected["role_slug"] = iamcore.AdminRoleSlug
		expected["permissions"] = nil // Permission values are asserted separately; order is not contractual.
	}
	if len(fields) != len(expected) {
		t.Fatalf("unexpected public field set (including optional/internal fields): %s", payload)
	}
	for key, value := range expected {
		got, ok := fields[key]
		if !ok {
			t.Fatalf("missing public field %q: %s", key, payload)
		}
		if key == "permissions" {
			continue
		}
		want, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s = %s, want %s", key, got, want)
		}
	}
}

// Mirror the adapter used by application wiring without bypassing IAM resolution.
type integrationAuthUserLookup struct {
	resolver iamcore.AuthUserResolver
}

func (l integrationAuthUserLookup) GetAuthUser(publicID string) (*authctx.User, error) {
	return l.resolver.ResolveAuthUser(publicID)
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
