package helpers

import (
	"testing"

	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
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

	dbA := NewSQLiteDB(t, &iamcore.IAMRole{})
	dbB := NewSQLiteDB(t, &iamcore.IAMRole{})

	role := SeedRole(t, dbA, iamcore.AdminRoleSlug)
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
			if err := db.Model(&iamcore.IAMRole{}).Count(&count).Error; err != nil {
				t.Fatalf("count roles in %s: %v", tt.target, err)
			}
			if count != tt.expectedCount {
				t.Fatalf("expected %d role(s) in %s, got %d", tt.expectedCount, tt.target, count)
			}
		})
	}
}
