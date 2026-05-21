package users

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type routeFakeService struct{}

func (routeFakeService) List(tc tenant.TenantContext, params query.ListParams) ([]UserResponse, int64, error) {
	return nil, 0, nil
}
func (routeFakeService) GetByPublicID(tc tenant.TenantContext, publicID string) (*UserResponse, error) {
	return nil, nil
}
func (routeFakeService) Create(tc tenant.TenantContext, req CreateUserRequest) (*UserResponse, error) {
	return nil, nil
}
func (routeFakeService) Update(tc tenant.TenantContext, publicID string, actorPublicID string, req UpdateUserRequest) (*UserResponse, error) {
	return nil, nil
}

func TestRegisterUserRoutesReturnStandardAuthzEnvelopes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		user   *authctx.User
		status int
	}{
		{name: "unauthenticated user route returns 401 envelope", status: http.StatusUnauthorized},
		{name: "missing permission returns 403 envelope", user: &authctx.User{PublicID: "user1", RoleSlug: "user", Permissions: []string{"users.view"}}, status: http.StatusForbidden},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			if tt.user != nil {
				r.Use(func(c *gin.Context) {
					authctx.SetGin(c, tt.user)
					c.Next()
				})
			}
			Register(r.Group(""), NewHandler(routeFakeService{}, nil), middleware.RequirePermission)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users", nil))

			if w.Code != tt.status {
				t.Fatalf("expected status %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
			var resp response.APIResponse[any]
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("expected standard envelope: %v", err)
			}
			if resp.Success {
				t.Fatalf("expected error envelope, got %#v", resp)
			}
		})
	}
}
func (routeFakeService) Delete(tc tenant.TenantContext, publicID string) error { return nil }
func (routeFakeService) ChangePassword(tc tenant.TenantContext, publicID string, actorPublicID string, req ChangePasswordRequest) error {
	return nil
}
func (routeFakeService) ToggleStatus(tc tenant.TenantContext, publicID string, req UpdateStatusRequest) (*UserResponse, error) {
	return nil, nil
}

func TestRegisterAppliesUserPermissionGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		method   string
		path     string
		expected []string
	}{
		{name: "list users", method: http.MethodGet, path: "/users", expected: []string{"users.index"}},
		{name: "create user", method: http.MethodPost, path: "/users", expected: []string{"users.create"}},
		{name: "view user", method: http.MethodGet, path: "/users/user1", expected: []string{"users.view"}},
		{name: "update user", method: http.MethodPut, path: "/users/user1", expected: []string{"users.update", "users.change_role"}},
		{name: "delete user", method: http.MethodDelete, path: "/users/user1", expected: []string{"users.delete"}},
		{name: "change password", method: http.MethodPatch, path: "/users/user1/password", expected: []string{"users.update"}},
		{name: "toggle status", method: http.MethodPatch, path: "/users/user1/status", expected: []string{"users.update"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			r := gin.New()
			Register(r.Group(""), NewHandler(routeFakeService{}, nil), func(slug string) gin.HandlerFunc {
				return func(c *gin.Context) {
					got = append(got, slug)
					if len(got) == len(tt.expected) {
						c.Status(http.StatusNoContent)
						c.Abort()
						return
					}
					c.Next()
				}
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))

			if w.Code != http.StatusNoContent {
				t.Fatalf("expected route guard to abort with 204, got %d", w.Code)
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("expected guards %v, got %v", tt.expected, got)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Fatalf("expected guards %v, got %v", tt.expected, got)
				}
			}
		})
	}
}
