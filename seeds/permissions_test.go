package seeds

import (
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedPermissions(t *testing.T) {
	database := newSeedTestDB(t)

	t.Run("seeds explicit system permissions", func(t *testing.T) {
		if err := seedPermissions(database); err != nil {
			t.Fatalf("seedPermissions failed: %v", err)
		}

		var count int64
		if err := database.Model(&permissions.Permission{}).Count(&count).Error; err != nil {
			t.Fatalf("failed to count permissions: %v", err)
		}
		if want := int64(len(systemPermissions())); count != want {
			t.Fatalf("expected %d permissions, got %d", want, count)
		}

		var permission permissions.Permission
		if err := database.Where("slug = ?", "users.change_role").First(&permission).Error; err != nil {
			t.Fatalf("expected users.change_role permission: %v", err)
		}
		if permission.Module != "users" || permission.Action != permissions.ActionChangeRole || !permission.IsSystem || permission.DisplayOrder == 0 {
			t.Fatalf("unexpected permission fields: %+v", permission)
		}

		var readCount int64
		if err := database.Model(&permissions.Permission{}).Where("action = ? OR slug LIKE ?", "read", "%.read").Count(&readCount).Error; err != nil {
			t.Fatalf("failed to count read permissions: %v", err)
		}
		if readCount != 0 {
			t.Fatalf("expected no ambiguous read permissions, got %d", readCount)
		}

		for _, slug := range []string{"roles.create", "roles.update", "roles.delete"} {
			var rolePermission permissions.Permission
			if err := database.Where("slug = ?", slug).First(&rolePermission).Error; err != nil {
				t.Fatalf("expected %s permission: %v", slug, err)
			}
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		if err := seedPermissions(database); err != nil {
			t.Fatalf("seedPermissions failed on second run: %v", err)
		}
		var count int64
		if err := database.Model(&permissions.Permission{}).Count(&count).Error; err != nil {
			t.Fatalf("failed to count permissions: %v", err)
		}
		if want := int64(len(systemPermissions())); count != want {
			t.Fatalf("expected %d permissions after rerun, got %d", want, count)
		}
	})
}

func TestSeedRolePermissions(t *testing.T) {
	database := newSeedTestDB(t)
	if err := seedRoles(database); err != nil {
		t.Fatalf("seedRoles failed: %v", err)
	}
	if err := seedPermissions(database); err != nil {
		t.Fatalf("seedPermissions failed: %v", err)
	}

	if err := seedRolePermissions(database); err != nil {
		t.Fatalf("seedRolePermissions failed: %v", err)
	}
	if err := seedRolePermissions(database); err != nil {
		t.Fatalf("seedRolePermissions failed on rerun: %v", err)
	}

	var links []permissions.RolePermission
	if err := database.Find(&links).Error; err != nil {
		t.Fatalf("failed to load role permissions: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected no role_permissions rows, got %d", len(links))
	}

	var rootRole roles.Role
	if err := database.Where("slug = ?", roles.RootRoleSlug).First(&rootRole).Error; err != nil {
		t.Fatalf("failed to load root role: %v", err)
	}
}

func TestAdminPermissionSlugsExcludeRootOnlyCompanyMutations(t *testing.T) {
	rootOnlyCompanyMutations := []string{
		"companies.create",
		"companies.update",
		"companies.delete",
	}

	adminPermissions := make(map[string]struct{})
	for _, slug := range adminPermissionSlugs() {
		adminPermissions[slug] = struct{}{}
	}

	for _, slug := range rootOnlyCompanyMutations {
		if _, ok := adminPermissions[slug]; ok {
			t.Fatalf("admin permissions must not include root-only company mutation %q", slug)
		}
	}
}

func TestAdminPermissionSlugsIncludesRoleCRUD(t *testing.T) {
	adminPermissions := make(map[string]struct{})
	for _, slug := range adminPermissionSlugs() {
		adminPermissions[slug] = struct{}{}
	}

	for _, slug := range []string{"roles.create", "roles.update", "roles.delete"} {
		if _, ok := adminPermissions[slug]; !ok {
			t.Fatalf("expected admin permissions to include %q", slug)
		}
	}
}

func TestSystemPermissionsCatalog(t *testing.T) {
	for _, permission := range systemPermissions() {
		t.Run(permission.Slug, func(t *testing.T) {
			if strings.Contains(permission.Slug, ".read") || permission.Action == "read" {
				t.Fatalf("permission uses ambiguous read action: %+v", permission)
			}
			if permission.Slug != permission.Module+"."+permission.Action {
				t.Fatalf("permission slug does not match module.action: %+v", permission)
			}
			if !permission.IsSystem {
				t.Fatalf("expected system permission: %+v", permission)
			}
			if permission.DisplayOrder <= 0 {
				t.Fatalf("expected positive display order: %+v", permission)
			}
		})
	}
}

func newSeedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := database.AutoMigrate(&roles.Role{}, &permissions.Permission{}, &permissions.RolePermission{}); err != nil {
		t.Fatalf("failed to migrate seed test db: %v", err)
	}
	return database
}
