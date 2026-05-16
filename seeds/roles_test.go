package seeds

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedRoles(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	// Create roles table manually for in-memory test
	if err := db.AutoMigrate(&roles.Role{}); err != nil {
		t.Fatalf("failed to migrate roles table: %v", err)
	}

	t.Run("seeds all roles on first run", func(t *testing.T) {
		if err := seedRoles(db); err != nil {
			t.Fatalf("seedRoles failed: %v", err)
		}

		var count int64
		if err := db.Model(&roles.Role{}).Count(&count).Error; err != nil {
			t.Fatalf("failed to count roles: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 roles, got %d", count)
		}

		for _, name := range []string{"root", "admin", "user"} {
			var role roles.Role
			if err := db.Where("name = ?", name).First(&role).Error; err != nil {
				t.Errorf("expected role %s to exist: %v", name, err)
				continue
			}
			if !role.IsSystem {
				t.Errorf("expected role %s to be system role", name)
			}
			if role.PublicID == "" {
				t.Errorf("expected role %s to have a public_id", name)
			}
			if role.Slug == "" {
				t.Errorf("expected role %s to have a slug", name)
			}
			if role.Description == "" {
				t.Errorf("expected role %s to have a description", name)
			}
		}
	})

	t.Run("is idempotent on second run", func(t *testing.T) {
		if err := seedRoles(db); err != nil {
			t.Fatalf("seedRoles failed on second run: %v", err)
		}

		var count int64
		if err := db.Model(&roles.Role{}).Count(&count).Error; err != nil {
			t.Fatalf("failed to count roles: %v", err)
		}
		if count != 3 {
			t.Errorf("expected still 3 roles after re-run, got %d", count)
		}
	})
}
