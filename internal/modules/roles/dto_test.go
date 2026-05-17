package roles

import (
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestCreateRoleRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateRoleRequest
		wantErr bool
		field   string
		msg     string
	}{
		{
			name: "valid request",
			req:  CreateRoleRequest{Name: "Editor", Slug: "editor"},
		},
		{
			name:    "missing name",
			req:     CreateRoleRequest{Slug: "editor"},
			wantErr: true,
			field:   "name",
			msg:     messages.MsgRequired,
		},
		{
			name:    "name too short",
			req:     CreateRoleRequest{Name: "E", Slug: "editor"},
			wantErr: true,
			field:   "name",
			msg:     messages.MsgMinLength,
		},
		{
			name:    "missing slug",
			req:     CreateRoleRequest{Name: "Editor"},
			wantErr: true,
			field:   "slug",
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

func TestUpdateRoleRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateRoleRequest
		wantErr bool
		field   string
		msg     string
	}{
		{
			name: "valid request",
			req:  UpdateRoleRequest{Name: "Editor", Slug: "editor"},
		},
		{
			name:    "missing name",
			req:     UpdateRoleRequest{Slug: "editor"},
			wantErr: true,
			field:   "name",
			msg:     messages.MsgRequired,
		},
		{
			name:    "missing slug",
			req:     UpdateRoleRequest{Name: "Editor"},
			wantErr: true,
			field:   "slug",
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
				if e == tt.msg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected message %q for field %q, got %v", tt.msg, tt.field, fieldErrs)
			}
		})
	}
}
