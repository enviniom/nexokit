package roles

import "testing"

func TestReservedRoleIdentities(t *testing.T) {
	for _, tt := range []struct {
		name, slug string
		want       bool
	}{
		{" Root ", "custom", true},
		{"Custom", " ADMIN ", true},
		{"USER", "custom", true},
		{"Custom", "custom", false},
		{"root-like", "administrator", false},
		{"", "", false},
	} {
		t.Run(tt.name+"/"+tt.slug, func(t *testing.T) {
			if got := IsReservedRoleIdentity(tt.name, tt.slug); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
	if RootRoleSlug != "root" || AdminRoleSlug != "admin" || UserRoleSlug != "user" || len(ReservedSlugs) != 3 {
		t.Fatal("application role identities changed")
	}
}
