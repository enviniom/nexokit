package view_session

import (
	"encoding/json"
	userentity "github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeHandlerService struct{}

func (fakeHandlerService) View(current *authctx.User) (*userentity.SessionView, error) {
	return &userentity.SessionView{
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
		var resp response.APIResponse[userentity.MeResponse]
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
		var resp response.APIResponse[userentity.MeResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(resp.Data.Permissions) != 3 {
			t.Fatalf("expected 3 permissions, got %d", len(resp.Data.Permissions))
		}
	})

	company := uint(7)
	for _, tc := range []struct {
		name        string
		permissions []string
		want        string
		company     *uint
	}{
		{name: "nil permissions", want: "null"},
		{name: "empty permissions", permissions: []string{}, want: "[]", company: &company},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(fakeHandlerService{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			authctx.SetGin(c, &authctx.User{PublicID: "user-2", Name: "Bob", Email: "b@example.com", Role: "viewer", RoleSlug: "viewer", RoleID: 3, IsActive: true, Permissions: tc.permissions, CompanyID: tc.company})

			h.Handle(c)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}
			var resp response.APIResponse[userentity.MeResponse]
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(resp.Data.Permissions) != 0 {
				t.Fatalf("expected empty permissions, got %#v", resp.Data.Permissions)
			}
			var envelope struct {
				Data map[string]json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if string(envelope.Data["permissions"]) != tc.want {
				t.Fatalf("permissions = %s, want %s", envelope.Data["permissions"], tc.want)
			}
			wantCount := 10
			if tc.company != nil {
				wantCount++
				if string(envelope.Data["company_id"]) != "7" {
					t.Fatalf("company_id = %s", envelope.Data["company_id"])
				}
			}
			if len(envelope.Data) != wantCount {
				t.Fatalf("unexpected field set: %s", w.Body.String())
			}
			for _, key := range []string{"created_by", "updated_by", "password", "password_hash", "id"} {
				if _, ok := envelope.Data[key]; ok {
					t.Fatalf("unexpected field %q", key)
				}
			}
		})
	}
}
