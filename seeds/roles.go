package seeds

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"gorm.io/gorm"
)

// RolesSeed idempotently seeds root, admin, and user system roles.
func RolesSeed() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return seedRoles(database)
}

func seedRoles(database *gorm.DB) error {
	roleList := []roles.Role{
		{Name: "root", Slug: roles.RootRoleSlug, Description: "System root role", IsSystem: true},
		{Name: "admin", Slug: roles.AdminRoleSlug, Description: "System admin role", IsSystem: true},
		{Name: "user", Slug: roles.UserRoleSlug, Description: "System user role", IsSystem: true},
	}

	for i := range roleList {
		var existing roles.Role
		result := database.Where("name = ?", roleList[i].Name).First(&existing)
		if result.Error == nil {
			// Already exists, skip
			continue
		}
		if result.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check role %s: %w", roleList[i].Name, result.Error)
		}

		publicID, err := identity.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate public id for role %s: %w", roleList[i].Name, err)
		}
		roleList[i].PublicID = publicID

		if err := database.Create(&roleList[i]).Error; err != nil {
			return fmt.Errorf("failed to create role %s: %w", roleList[i].Name, err)
		}
	}

	return nil
}
