package core

import "testing"

// TestIAMTableNames ensures IAM partial models map to the real migration/legacy
// table names, not the GORM default pluralized form (e.g. "iam_users").
// A regression here causes silent runtime failures against production databases
// while tests pass because AutoMigrate creates the wrong tables in SQLite.
func TestIAMTableNames(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "IAMUser", got: IAMUser{}.TableName(), want: "users"},
		{name: "IAMRole", got: IAMRole{}.TableName(), want: "roles"},
		{name: "IAMPermission", got: IAMPermission{}.TableName(), want: "permissions"},
		{name: "IAMCompany", got: IAMCompany{}.TableName(), want: "companies"},
		{name: "IAMRolePermission", got: IAMRolePermission{}.TableName(), want: "role_permissions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("expected table %q, got %q", tt.want, tt.got)
			}
		})
	}
}
