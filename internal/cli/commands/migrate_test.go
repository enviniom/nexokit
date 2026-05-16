package commands

import "testing"

func TestIsValidMigrationName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple snake", "create_users_table", true},
		{"with digits", "add_v2_columns", true},
		{"single word", "users", true},
		{"empty", "", false},
		{"uppercase", "CreateUsers", false},
		{"hyphen", "create-users", false},
		{"space", "create users", false},
		{"leading underscore", "_users", true},
		{"trailing underscore", "users_", true},
		{"special char", "users@table", false},
		{"mixed case", "Create_Users", false},
		{"only digits", "123", true},
		{"underscore only", "___", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMigrationName(tt.input)
			if got != tt.want {
				t.Errorf("isValidMigrationName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
