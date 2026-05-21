package helpers

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/roles"
)

func TestNewSQLiteDB_IsolatedPerCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		target        string
		expectedCount int64
	}{
		{name: "seeded database contains role", target: "dbA", expectedCount: 1},
		{name: "separate database stays empty", target: "dbB", expectedCount: 0},
	}

	dbA := NewSQLiteDB(t, &roles.Role{})
	dbB := NewSQLiteDB(t, &roles.Role{})

	role := SeedRole(t, dbA, roles.AdminRoleSlug)
	if role.ID == 0 {
		t.Fatalf("expected role id to be persisted")
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := dbA
			if tt.target == "dbB" {
				db = dbB
			}

			var count int64
			if err := db.Model(&roles.Role{}).Count(&count).Error; err != nil {
				t.Fatalf("count roles in %s: %v", tt.target, err)
			}
			if count != tt.expectedCount {
				t.Fatalf("expected %d role(s) in %s, got %d", tt.expectedCount, tt.target, count)
			}
		})
	}
}
