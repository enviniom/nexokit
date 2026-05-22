package seeds

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	"gorm.io/gorm"
)

// RolePermissionsSeed seeds the permissions catalog only.
func RolePermissionsSeed() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := seedPermissions(database); err != nil {
		return err
	}
	return nil
}

func seedRolePermissions(database *gorm.DB) error {
	// Root receives "*" via middleware (authctx.AttachPermissions), so no role_permissions rows are seeded.
	return nil
}
