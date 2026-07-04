package str

import "testing"

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"leading and trailing whitespace", "  Example.COM  ", "example.com"},
		{"mixed case", "AcMe.Co", "acme.co"},
		{"already lower", "example.com", "example.com"},
		{"trailing dot removed", "Example.COM.", "example.com"},
		{"internal dot preserved", "sub.Example.COM", "sub.example.com"},
		{"internal control char preserved", "A\tB.com", "a\tb.com"},
		{"control char at edge trimmed", "\tExample.COM\n", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDomain(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeDomain(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
