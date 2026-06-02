package core

import "testing"

func TestIsReservedRoleIdentity(t *testing.T) {
	tests := []struct {
		name  string
		role  string
		slug  string
		wants bool
	}{
		{name: "reserved by name", role: "Root", slug: "custom", wants: true},
		{name: "reserved by slug", role: "Custom", slug: "admin", wants: true},
		{name: "reserved with spaces and case", role: "  USER ", slug: "custom", wants: true},
		{name: "non reserved", role: "Manager", slug: "manager", wants: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsReservedRoleIdentity(tt.role, tt.slug)
			if got != tt.wants {
				t.Fatalf("expected %v, got %v", tt.wants, got)
			}
		})
	}
}
