package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/token"
	"github.com/gin-gonic/gin"
)

type fakeAccessParser struct {
	claims *token.AccessClaims
	err    error
}

func (f fakeAccessParser) ParseAccess(token string) (*token.AccessClaims, error) {
	return f.claims, f.err
}

type fakeAuthUserLookup struct {
	user *authctx.User
	err  error
}

func (f fakeAuthUserLookup) GetAuthUser(publicID string) (*authctx.User, error) {
	return f.user, f.err
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows valid token and injects active user", func(t *testing.T) {
		r := gin.New()
		r.Use(Auth(fakeAccessParser{claims: &token.AccessClaims{Sub: "user1", TokenType: "access", ExpiresAt: time.Now().Add(time.Hour)}}, fakeAuthUserLookup{user: &authctx.User{PublicID: "user1", Email: "user@example.com", IsActive: true}}))
		r.GET("/protected", func(c *gin.Context) {
			user, ok := authctx.FromGin(c)
			if !ok {
				c.Status(http.StatusInternalServerError)
				return
			}
			response.Success(c, "ok", gin.H{"public_id": user.PublicID, "email": user.Email})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if !json.Valid(w.Body.Bytes()) || strings.Contains(w.Body.String(), "password") {
			t.Fatalf("expected valid JSON without password fields, got %s", w.Body.String())
		}
	})

	cases := []struct {
		name   string
		header string
		parser fakeAccessParser
		lookup fakeAuthUserLookup
		status int
	}{
		{name: "missing bearer token", status: http.StatusUnauthorized},
		{name: "expired token", header: "Bearer expired", parser: fakeAccessParser{err: errors.New("expired")}, status: http.StatusUnauthorized},
		{name: "inactive user", header: "Bearer valid", parser: fakeAccessParser{claims: &token.AccessClaims{Sub: "user1", TokenType: "access", ExpiresAt: time.Now().Add(time.Hour)}}, lookup: fakeAuthUserLookup{user: &authctx.User{PublicID: "user1", IsActive: false}}, status: http.StatusUnauthorized},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(Auth(tt.parser, tt.lookup))
			r.GET("/protected", func(c *gin.Context) { response.Success[any](c, "ok", nil) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, w.Code)
			}
			var resp response.APIResponse[any]
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("expected standard JSON envelope: %v", err)
			}
			if resp.Success {
				t.Fatal("expected error response")
			}
		})
	}
}
