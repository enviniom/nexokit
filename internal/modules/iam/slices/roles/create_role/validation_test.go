package create_role

import (
	"fmt"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"strings"
	"testing"
)

// Relocated from core: both role requests retain the same validation contract.
func TestRoleRequestsValidate(t *testing.T) {
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			cases := []struct {
				name, roleName, slug string
				want                 map[string]string
			}{
				{"valid request", "Admin", "admin-custom", nil},
				{"all required fields missing", "", "", map[string]string{"name": messages.MsgRequired, "slug": messages.MsgRequired}},
				{"name too short", "A", "admin-custom", map[string]string{"name": fmt.Sprintf(messages.MsgMinLength, 2)}},
				{"slug invalid", "Admin", "invalid slug", map[string]string{"slug": messages.MsgValidSlug}},
				{"underscore slug invalid", "Admin", "invalid_slug", map[string]string{"slug": messages.MsgValidSlug}},
			}
			for _, tt := range cases {
				t.Run(tt.name, func(t *testing.T) {
					errs := (iamrole.CreateRoleRequest{Name: tt.roleName, Slug: tt.slug, Description: "Custom admin role"}).Validate()
					if operation == "update" {
						errs = (iamrole.UpdateRoleRequest{Name: tt.roleName, Slug: tt.slug, Description: "Custom admin role"}).Validate()
					}
					if len(tt.want) == 0 && len(errs) != 0 {
						t.Fatalf("unexpected errors: %v", errs)
					}
					for field, msg := range tt.want {
						if !strings.Contains(strings.Join(errs[field], "; "), msg) {
							t.Errorf("%s: want %q, got %v", field, msg, errs[field])
						}
					}
				})
			}
		})
	}
}

func TestAssignRolePermissionsRequestValidate(t *testing.T) {
	for _, tt := range []struct {
		name        string
		permissions []string
		invalid     bool
	}{
		{"valid empty slice", []string{}, false},
		{"valid with permissions", []string{"users.read", "roles.read"}, false},
		{"nil permissions", nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			errs := (iamrole.AssignRolePermissionsRequest{Permissions: tt.permissions}).Validate()
			if tt.invalid {
				if !strings.Contains(strings.Join(errs["permissions"], "; "), "permissions is required") {
					t.Fatalf("unexpected errors: %v", errs)
				}
			} else if len(errs) != 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
	}
}
