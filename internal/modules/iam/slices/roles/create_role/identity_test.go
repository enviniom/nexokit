package create_role

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"testing"
)

func TestIsReservedRoleIdentity(t *testing.T) {
	tests := []struct {
		name, role, slug string
		wants            bool
	}{
		{"reserved by name", "Root", "custom", true},
		{"reserved by slug", "Custom", "admin", true},
		{"reserved with spaces and case", "  USER ", "custom", true},
		{"non reserved", "Manager", "manager", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.IsReservedRoleIdentity(tt.role, tt.slug); got != tt.wants {
				t.Fatalf("expected %v, got %v", tt.wants, got)
			}
		})
	}
}

func TestIAMRoleTableNames(t *testing.T) {
	tests := []struct{ name, got, want string }{
		{"IAMRole", iamrole.IAMRole{}.TableName(), "roles"},
		{"IAMCompany", iamrole.IAMCompany{}.TableName(), "companies"},
		{"IAMRolePermission", iamrole.IAMRolePermission{}.TableName(), "role_permissions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("expected table %q, got %q", tt.want, tt.got)
			}
		})
	}
}
