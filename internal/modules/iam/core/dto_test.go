package core

import (
	"fmt"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestCreateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateUserRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: CreateUserRequest{
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "password123",
				RoleID:   2,
			},
		},
		{
			name: "all required fields missing",
			req:  CreateUserRequest{},
			wantMsgs: map[string]string{
				"name":     messages.MsgRequired,
				"email":    messages.MsgRequired,
				"password": messages.MsgRequired,
				"role_id":  messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: CreateUserRequest{
				Name:     "A",
				Email:    "alice@example.com",
				Password: "password123",
				RoleID:   2,
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "email invalid",
			req: CreateUserRequest{
				Name:     "Alice",
				Email:    "not-an-email",
				Password: "password123",
				RoleID:   2,
			},
			wantMsgs: map[string]string{
				"email": messages.MsgValidEmail,
			},
		},
		{
			name: "password too short",
			req: CreateUserRequest{
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "short",
				RoleID:   2,
			},
			wantMsgs: map[string]string{
				"password": fmt.Sprintf(messages.MsgMinLength, 8),
			},
		},
		{
			name: "multiple validation failures",
			req: CreateUserRequest{
				Name:     "",
				Email:    "bad-email",
				Password: "tiny",
				RoleID:   0,
			},
			wantMsgs: map[string]string{
				"name":     messages.MsgRequired,
				"email":    messages.MsgValidEmail,
				"password": fmt.Sprintf(messages.MsgMinLength, 8),
				"role_id":  messages.MsgRequired,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestUpdateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      UpdateUserRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: UpdateUserRequest{
				Name:  "Alice",
				Email: "alice@example.com",
			},
		},
		{
			name: "all required fields missing",
			req:  UpdateUserRequest{},
			wantMsgs: map[string]string{
				"name":  messages.MsgRequired,
				"email": messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: UpdateUserRequest{
				Name:  "A",
				Email: "alice@example.com",
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "email invalid",
			req: UpdateUserRequest{
				Name:  "Alice",
				Email: "not-an-email",
			},
			wantMsgs: map[string]string{
				"email": messages.MsgValidEmail,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestChangeUserRoleRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      ChangeUserRoleRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req:  ChangeUserRoleRequest{RoleID: "role-admin"},
		},
		{
			name: "missing role_id",
			req:  ChangeUserRoleRequest{},
			wantMsgs: map[string]string{
				"role_id": messages.MsgRequired,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestChangePasswordRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      ChangePasswordRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: ChangePasswordRequest{
				CurrentPassword: "oldpassword123",
				NewPassword:     "newpassword123",
			},
		},
		{
			name: "all required fields missing",
			req:  ChangePasswordRequest{},
			wantMsgs: map[string]string{
				"current_password": messages.MsgRequired,
				"new_password":     messages.MsgRequired,
			},
		},
		{
			name: "new password too short",
			req: ChangePasswordRequest{
				CurrentPassword: "oldpassword123",
				NewPassword:     "short",
			},
			wantMsgs: map[string]string{
				"new_password": fmt.Sprintf(messages.MsgMinLength, 8),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestUpdateStatusRequest_Validate(t *testing.T) {
	tests := []struct {
		name string
		req  UpdateStatusRequest
	}{
		{name: "active true", req: UpdateStatusRequest{IsActive: true}},
		{name: "active false", req: UpdateStatusRequest{IsActive: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if len(errs) > 0 {
				t.Fatalf("expected no validation errors, got %v", errs)
			}
		})
	}
}

func TestAssignRolePermissionsRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      AssignRolePermissionsRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid empty slice",
			req:  AssignRolePermissionsRequest{Permissions: []string{}},
		},
		{
			name: "valid with permissions",
			req:  AssignRolePermissionsRequest{Permissions: []string{"users.read", "roles.read"}},
		},
		{
			name: "nil permissions",
			req:  AssignRolePermissionsRequest{},
			wantMsgs: map[string]string{
				"permissions": "permissions is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestCreateRoleRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateRoleRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: CreateRoleRequest{
				Name:        "Admin",
				Slug:        "admin-custom",
				Description: "Custom admin role",
			},
		},
		{
			name: "all required fields missing",
			req:  CreateRoleRequest{},
			wantMsgs: map[string]string{
				"name": messages.MsgRequired,
				"slug": messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: CreateRoleRequest{
				Name:        "A",
				Slug:        "admin-custom",
				Description: "Custom admin role",
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "slug invalid",
			req: CreateRoleRequest{
				Name:        "Admin",
				Slug:        "invalid slug",
				Description: "Custom admin role",
			},
			wantMsgs: map[string]string{
				"slug": messages.MsgValidSlug,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestUpdateRoleRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      UpdateRoleRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: UpdateRoleRequest{
				Name:        "Admin",
				Slug:        "admin-custom",
				Description: "Custom admin role",
			},
		},
		{
			name: "all required fields missing",
			req:  UpdateRoleRequest{},
			wantMsgs: map[string]string{
				"name": messages.MsgRequired,
				"slug": messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: UpdateRoleRequest{
				Name:        "A",
				Slug:        "admin-custom",
				Description: "Custom admin role",
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "slug invalid",
			req: UpdateRoleRequest{
				Name:        "Admin",
				Slug:        "invalid_slug",
				Description: "Custom admin role",
			},
			wantMsgs: map[string]string{
				"slug": messages.MsgValidSlug,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestUpdatePermissionRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      UpdatePermissionRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req:  UpdatePermissionRequest{Name: "Users Read"},
		},
		{
			name: "missing name",
			req:  UpdatePermissionRequest{},
			wantMsgs: map[string]string{
				"name": messages.MsgRequired,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func assertValidation(t *testing.T, errs map[string][]string, wantMsgs map[string]string) {
	t.Helper()

	if len(wantMsgs) == 0 {
		if len(errs) > 0 {
			t.Fatalf("expected no validation errors, got %v", errs)
		}
		return
	}

	if len(errs) == 0 {
		t.Fatalf("expected validation errors, got none")
	}

	for field, wantMsg := range wantMsgs {
		fieldErrs, ok := errs[field]
		if !ok {
			t.Errorf("expected error for field %q, got none", field)
			continue
		}
		if len(fieldErrs) == 0 {
			t.Errorf("expected non-empty error list for field %q", field)
			continue
		}

		joined := joinErrors(fieldErrs)
		if !contains(joined, wantMsg) {
			t.Errorf("field %q: expected message containing %q, got %q", field, wantMsg, joined)
		}
	}
}

func joinErrors(errs []string) string {
	s := ""
	for i, e := range errs {
		if i > 0 {
			s += "; "
		}
		s += e
	}
	return s
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(haystack) > 0 && containsSubstring(haystack, needle))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
