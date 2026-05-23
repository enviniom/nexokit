package users

import (
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestCreateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateUserRequest
		wantErr bool
		field   string
		msg     string
	}{
		{
			name: "valid request with company",
			req:  CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(1)},
		},
		{
			name: "any role can structurally omit company",
			req:  CreateUserRequest{Name: "Root", Email: "root@example.com", Password: "Password1", RoleID: RootRoleID},
		},
		{
			name:    "missing name",
			req:     CreateUserRequest{Email: "alice@example.com", Password: "Password1", RoleID: 2},
			wantErr: true,
			field:   "name",
			msg:     messages.MsgRequired,
		},
		{
			name:    "name too short",
			req:     CreateUserRequest{Name: "A", Email: "alice@example.com", Password: "Password1", RoleID: 2},
			wantErr: true,
			field:   "name",
			msg:     messages.MsgMinLength,
		},
		{
			name:    "missing email",
			req:     CreateUserRequest{Name: "Alice", Password: "Password1", RoleID: 2},
			wantErr: true,
			field:   "email",
			msg:     messages.MsgRequired,
		},
		{
			name:    "invalid email",
			req:     CreateUserRequest{Name: "Alice", Email: "not-an-email", Password: "Password1", RoleID: 2},
			wantErr: true,
			field:   "email",
			msg:     messages.MsgValidEmail,
		},
		{
			name:    "missing password",
			req:     CreateUserRequest{Name: "Alice", Email: "alice@example.com", RoleID: 2},
			wantErr: true,
			field:   "password",
			msg:     messages.MsgRequired,
		},
		{
			name:    "password too short",
			req:     CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "short", RoleID: 2},
			wantErr: true,
			field:   "password",
			msg:     messages.MsgMinLength,
		},
		{
			name:    "missing role_id",
			req:     CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1"},
			wantErr: true,
			field:   "role_id",
			msg:     messages.MsgRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if !tt.wantErr {
				if errs.HasErrors() {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if !errs.HasErrors() {
				t.Fatal("expected validation errors, got none")
			}
			fieldErrs, ok := errs[tt.field]
			if !ok {
				t.Fatalf("expected error on field %q, got errors on other fields: %v", tt.field, errs)
			}
			found := false
			for _, e := range fieldErrs {
				if strings.Contains(tt.msg, "%d") {
					if strings.HasPrefix(e, strings.Split(tt.msg, "%d")[0]) {
						found = true
						break
					}
				} else if e == tt.msg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected message matching %q for field %q, got %v", tt.msg, tt.field, fieldErrs)
			}
		})
	}
}

func TestUpdateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateUserRequest
		wantErr bool
		field   string
		msg     string
	}{
		{
			name: "valid request with company",
			req:  UpdateUserRequest{Name: "Alice", Email: "alice@example.com", CompanyID: uintPtr(1)},
		},
		{
			name: "any role can structurally omit company",
			req:  UpdateUserRequest{Name: "Root", Email: "root@example.com"},
		},
		{
			name:    "missing name",
			req:     UpdateUserRequest{Email: "alice@example.com"},
			wantErr: true,
			field:   "name",
			msg:     messages.MsgRequired,
		},
		{
			name:    "missing email",
			req:     UpdateUserRequest{Name: "Alice"},
			wantErr: true,
			field:   "email",
			msg:     messages.MsgRequired,
		},
		{
			name:    "invalid email",
			req:     UpdateUserRequest{Name: "Alice", Email: "bad"},
			wantErr: true,
			field:   "email",
			msg:     messages.MsgValidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if !tt.wantErr {
				if errs.HasErrors() {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if !errs.HasErrors() {
				t.Fatal("expected validation errors, got none")
			}
			fieldErrs, ok := errs[tt.field]
			if !ok {
				t.Fatalf("expected error on field %q, got errors on other fields: %v", tt.field, errs)
			}
			found := false
			for _, e := range fieldErrs {
				if strings.Contains(tt.msg, "%d") {
					if strings.HasPrefix(e, strings.Split(tt.msg, "%d")[0]) {
						found = true
						break
					}
				} else if e == tt.msg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected message matching %q for field %q, got %v", tt.msg, tt.field, fieldErrs)
			}
		})
	}
}

func TestChangePasswordRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ChangePasswordRequest
		wantErr bool
		field   string
		msg     string
	}{
		{
			name: "valid request",
			req:  ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpassword"},
		},
		{
			name:    "missing current_password",
			req:     ChangePasswordRequest{NewPassword: "newpassword"},
			wantErr: true,
			field:   "current_password",
			msg:     messages.MsgRequired,
		},
		{
			name:    "missing new_password",
			req:     ChangePasswordRequest{CurrentPassword: "oldpass"},
			wantErr: true,
			field:   "new_password",
			msg:     messages.MsgRequired,
		},
		{
			name:    "new_password too short",
			req:     ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "short"},
			wantErr: true,
			field:   "new_password",
			msg:     messages.MsgMinLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if !tt.wantErr {
				if errs.HasErrors() {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if !errs.HasErrors() {
				t.Fatal("expected validation errors, got none")
			}
			fieldErrs, ok := errs[tt.field]
			if !ok {
				t.Fatalf("expected error on field %q, got errors on other fields: %v", tt.field, errs)
			}
			found := false
			for _, e := range fieldErrs {
				if strings.Contains(tt.msg, "%d") {
					if strings.HasPrefix(e, strings.Split(tt.msg, "%d")[0]) {
						found = true
						break
					}
				} else if e == tt.msg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected message matching %q for field %q, got %v", tt.msg, tt.field, fieldErrs)
			}
		})
	}
}

func TestUpdateStatusRequest_Validate(t *testing.T) {
	// UpdateStatusRequest has no fields to validate beyond the boolean,
	// so Validate should always return empty errors.
	req := UpdateStatusRequest{IsActive: false}
	errs := req.Validate()
	if errs.HasErrors() {
		t.Fatalf("expected no errors for UpdateStatusRequest, got %v", errs)
	}
}
