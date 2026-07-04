package str

import "testing"

func TestNormalizeSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"leading and trailing whitespace", "  HelloWorld  ", "helloworld"},
		{"mixed case", "AcMe-Corp", "acme-corp"},
		{"already lower", "lowercase", "lowercase"},
		{"trailing dot kept", "Slug.", "slug."},
		{"internal control char preserved", "A\tB", "a\tb"},
		{"control char at edge trimmed", "\tHello\n", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSlug(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeSlug(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
