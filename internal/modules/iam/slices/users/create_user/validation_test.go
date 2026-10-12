package create_user

import (
	"fmt"
	"testing"

	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestCreateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      iamuser.CreateUserRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: iamuser.CreateUserRequest{
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "password123",
				RoleID:   2,
			},
		},
		{
			name: "all required fields missing",
			req:  iamuser.CreateUserRequest{},
			wantMsgs: map[string]string{
				"name":     messages.MsgRequired,
				"email":    messages.MsgRequired,
				"password": messages.MsgRequired,
				"role_id":  messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: iamuser.CreateUserRequest{
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
			req: iamuser.CreateUserRequest{
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
			req: iamuser.CreateUserRequest{
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
			req: iamuser.CreateUserRequest{
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
		req      iamuser.UpdateUserRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: iamuser.UpdateUserRequest{
				Name:  "Alice",
				Email: "alice@example.com",
			},
		},
		{
			name: "all required fields missing",
			req:  iamuser.UpdateUserRequest{},
			wantMsgs: map[string]string{
				"name":  messages.MsgRequired,
				"email": messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: iamuser.UpdateUserRequest{
				Name:  "A",
				Email: "alice@example.com",
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "email invalid",
			req: iamuser.UpdateUserRequest{
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
		req      iamuser.ChangeUserRoleRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req:  iamuser.ChangeUserRoleRequest{RoleID: "role-admin"},
		},
		{
			name: "missing role_id",
			req:  iamuser.ChangeUserRoleRequest{},
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
		req      iamuser.ChangePasswordRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: iamuser.ChangePasswordRequest{
				CurrentPassword: "oldpassword123",
				NewPassword:     "newpassword123",
			},
		},
		{
			name: "all required fields missing",
			req:  iamuser.ChangePasswordRequest{},
			wantMsgs: map[string]string{
				"current_password": messages.MsgRequired,
				"new_password":     messages.MsgRequired,
			},
		},
		{
			name: "new password too short",
			req: iamuser.ChangePasswordRequest{
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
		req  iamuser.UpdateStatusRequest
	}{
		{name: "active true", req: iamuser.UpdateStatusRequest{IsActive: true}},
		{name: "active false", req: iamuser.UpdateStatusRequest{IsActive: false}},
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
