package seeds

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"gorm.io/gorm"
)

// PermissionsSeed idempotently seeds base system permissions.
func PermissionsSeed() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return seedPermissions(database)
}

func seedPermissions(database *gorm.DB) error {
	permissionList := systemPermissions()
	for i := range permissionList {
		var existing permissions.Permission
		result := database.Where("slug = ?", permissionList[i].Slug).First(&existing)
		if result.Error == nil {
			continue
		}
		if result.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check permission %s: %w", permissionList[i].Slug, result.Error)
		}

		publicID, err := identity.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate public id for permission %s: %w", permissionList[i].Slug, err)
		}
		permissionList[i].PublicID = publicID
		permissionList[i].IsSystem = true

		if err := database.Create(&permissionList[i]).Error; err != nil {
			return fmt.Errorf("failed to create permission %s: %w", permissionList[i].Slug, err)
		}
	}
	return nil
}

func systemPermissions() []permissions.Permission {
	return []permissions.Permission{
		permission("users", permissions.ActionIndex, "List users", "Allows listing users", 10),
		permission("users", permissions.ActionView, "View users", "Allows viewing user details", 20),
		permission("users", permissions.ActionCreate, "Create users", "Allows creating users", 30),
		permission("users", permissions.ActionUpdate, "Update users", "Allows updating users", 40),
		permission("users", permissions.ActionDelete, "Delete users", "Allows deleting users", 50),
		permission("users", permissions.ActionChangeRole, "Change user roles", "Allows changing a user's role", 60),

		permission("roles", permissions.ActionIndex, "List roles", "Allows listing roles", 10),
		permission("roles", permissions.ActionView, "View roles", "Allows viewing role details", 20),
		permission("roles", permissions.ActionCreate, "Create roles", "Allows creating roles", 30),
		permission("roles", permissions.ActionUpdate, "Update roles", "Allows updating roles", 40),
		permission("roles", permissions.ActionDelete, "Delete roles", "Allows deleting roles", 50),
		permission("roles", permissions.ActionAssignPermissions, "Assign role permissions", "Allows assigning permissions to roles", 60),

		permission("companies", permissions.ActionIndex, "List companies", "Allows listing companies", 10),
		permission("companies", permissions.ActionView, "View companies", "Allows viewing company details", 20),
		permission("companies", permissions.ActionCreate, "Create companies", "Allows creating companies", 30),
		permission("companies", permissions.ActionUpdate, "Update companies", "Allows updating companies", 40),
		permission("companies", permissions.ActionDelete, "Delete companies", "Allows deleting companies", 50),

		permission("settings", permissions.ActionView, "View settings", "Allows viewing settings", 10),
		permission("settings", permissions.ActionUpdate, "Update settings", "Allows updating settings", 20),

		permission("auth", permissions.ActionView, "View auth profile", "Allows viewing the authenticated profile", 10),
		permission("permissions", permissions.ActionManage, "Manage permissions", "Allows managing permissions", 10),
	}
}

func permission(module, action, name, description string, order int) permissions.Permission {
	return permissions.Permission{
		Slug:         fmt.Sprintf("%s.%s", module, action),
		Name:         name,
		Module:       module,
		Action:       action,
		Description:  description,
		IsSystem:     true,
		DisplayOrder: order,
	}
}
