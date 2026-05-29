package view_session

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct{}

func (fakeHandlerService) View(current *authctx.User) (*SessionView, error) {
	return &SessionView{
		PublicID:    current.PublicID,
		Name:        current.Name,
		Email:       current.Email,
		IsActive:    current.IsActive,
		RoleID:      current.RoleID,
		RoleName:    current.Role,
		RoleSlug:    current.RoleSlug,
		CompanyID:   current.CompanyID,
		Permissions: current.Permissions,
	}, nil
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns authenticated session payload", func(t *testing.T) {
		h := NewHandler(fakeHandlerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		authctx.SetGin(c, &authctx.User{PublicID: "user-1", Name: "Alice", Email: "a@example.com", Role: "admin", RoleSlug: "admin", RoleID: 2, IsActive: true, Permissions: []string{"users.list"}})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		var resp response.APIResponse[core.MeResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.Data.PublicID != "user-1" || resp.Data.RoleSlug != "admin" {
			t.Fatalf("unexpected me payload: %#v", resp.Data)
		}
	})

	t.Run("returns all permissions for root user", func(t *testing.T) {
		h := NewHandler(fakeHandlerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		authctx.SetGin(c, &authctx.User{PublicID: "root-1", Name: "Root", Email: "root@example.com", Role: "root", RoleSlug: "root", RoleID: 1, IsRoot: true, IsActive: true, Permissions: []string{"users.list", "users.create", "roles.manage"}})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		var resp response.APIResponse[core.MeResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(resp.Data.Permissions) != 3 {
			t.Fatalf("expected 3 permissions, got %d", len(resp.Data.Permissions))
		}
	})

	t.Run("returns empty permissions when none assigned", func(t *testing.T) {
		h := NewHandler(fakeHandlerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		authctx.SetGin(c, &authctx.User{PublicID: "user-2", Name: "Bob", Email: "b@example.com", Role: "viewer", RoleSlug: "viewer", RoleID: 3, IsActive: true, Permissions: []string{}})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		var resp response.APIResponse[core.MeResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(resp.Data.Permissions) != 0 {
			t.Fatalf("expected empty permissions, got %#v", resp.Data.Permissions)
		}
	})
}
