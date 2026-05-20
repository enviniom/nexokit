package seeds

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"gorm.io/gorm"
)

// RolePermissionsSeed idempotently assigns base permissions to system roles.
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
	return seedRolePermissions(database)
}

func seedRolePermissions(database *gorm.DB) error {
	assignments := map[string][]string{
		roles.RootRoleSlug:  allSystemPermissionSlugs(),
		roles.AdminRoleSlug: adminPermissionSlugs(),
		roles.UserRoleSlug:  userPermissionSlugs(),
	}

	for roleSlug, permissionSlugs := range assignments {
		var role roles.Role
		if err := database.Where("slug = ?", roleSlug).First(&role).Error; err != nil {
			return fmt.Errorf("failed to load role %s: %w", roleSlug, err)
		}

		for _, permissionSlug := range permissionSlugs {
			var permission permissions.Permission
			if err := database.Where("slug = ?", permissionSlug).First(&permission).Error; err != nil {
				return fmt.Errorf("failed to load permission %s: %w", permissionSlug, err)
			}

			link := permissions.RolePermission{RoleID: role.ID, PermissionID: permission.ID}
			result := database.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).First(&permissions.RolePermission{})
			if result.Error == nil {
				continue
			}
			if result.Error != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to check role permission %s/%s: %w", roleSlug, permissionSlug, result.Error)
			}
			if err := database.Create(&link).Error; err != nil {
				return fmt.Errorf("failed to assign permission %s to role %s: %w", permissionSlug, roleSlug, err)
			}
		}
	}

	return nil
}

func allSystemPermissionSlugs() []string {
	items := systemPermissions()
	slugs := make([]string, len(items))
	for i := range items {
		slugs[i] = items[i].Slug
	}
	return slugs
}

func adminPermissionSlugs() []string {
	return []string{
		"users.index",
		"users.view",
		"users.create",
		"users.update",
		"users.delete",
		"users.change_role",
		"roles.index",
		"roles.view",
		"roles.assign_permissions",
		"companies.index",
		"companies.view",
		"settings.view",
		"settings.update",
		"auth.view",
		"permissions.manage",
	}
}

func userPermissionSlugs() []string {
	return []string{"users.index", "users.view", "auth.view"}
}
