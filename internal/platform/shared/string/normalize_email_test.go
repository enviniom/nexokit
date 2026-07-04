package str

import "testing"

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"leading and trailing whitespace", "  User@Example.COM  ", "user@example.com"},
		{"mixed case local part", "USER@example.com", "user@example.com"},
		{"mixed case domain part", "user@EXAMPLE.com", "user@example.com"},
		{"already lower", "user@example.com", "user@example.com"},
		{"trailing dot kept", "User@Example.com.", "user@example.com."},
		{"internal control char preserved", "A\tB@example.com", "a\tb@example.com"},
		{"control char at edge trimmed", "\tUser@Example.COM\n", "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeEmail(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeEmail(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
