package assign_role_to_user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// --- fake cache ---

type fakeCache struct {
	deletedKeys []string
	deleteErr   error
}

func (f *fakeCache) Get(context.Context, string) ([]byte, error) { return nil, cache.ErrCacheMiss }
func (f *fakeCache) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (f *fakeCache) Delete(_ context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return f.deleteErr
}
func (f *fakeCache) Exists(context.Context, string) (bool, error) { return false, nil }
func (f *fakeCache) Close() error                                 { return nil }

// --- seed helper ---

func seedDB(t *testing.T, db *gorm.DB) (role core.IAMRole, user core.IAMUser) {
	t.Helper()
	if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role = core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-admin"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	rootRole := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-root"}, Name: "Root", Slug: "root", IsSystem: true}
	if err := db.Create(&rootRole).Error; err != nil {
		t.Fatalf("seed root role: %v", err)
	}

	companyID := uint(42)
	user = core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		CompanyID:    &companyID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return role, user
}

// --- tests ---

func TestRepositoryGetUserByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, _ = seedDB(t, db)

	repo := NewRepository(db, nil)

	t.Run("returns user with role preloaded", func(t *testing.T) {
		user, err := repo.GetUserByPublicID(tenant.NewRoot(), "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.PublicID != "user-1" {
			t.Fatalf("expected user-1, got %s", user.PublicID)
		}
		if user.Role.Name != "Admin" {
			t.Fatalf("expected role Admin, got %s", user.Role.Name)
		}
	})

	t.Run("returns ErrNotFound for missing user", func(t *testing.T) {
		_, err := repo.GetUserByPublicID(tenant.NewRoot(), "missing")
		if !errors.Is(err, core.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRepositoryGetRoleBySlug(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, _ = seedDB(t, db)

	repo := NewRepository(db, nil)

	t.Run("returns role when slug exists", func(t *testing.T) {
		role, err := repo.GetRoleBySlug("root")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if role.Slug != "root" {
			t.Fatalf("expected slug root, got %s", role.Slug)
		}
	})

	t.Run("returns ErrNotFound for missing slug", func(t *testing.T) {
		_, err := repo.GetRoleBySlug("nonexistent")
		if !errors.Is(err, core.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRepositoryGetAssignableRoleByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, _ = seedDB(t, db)

	repo := NewRepository(db, nil)

	t.Run("returns role when public ID exists", func(t *testing.T) {
		role, err := repo.GetAssignableRoleByPublicID(tenant.NewRoot(), "role-admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if role.Slug != "admin" {
			t.Fatalf("expected slug admin, got %s", role.Slug)
		}
	})

	t.Run("returns ErrNotFound for missing role", func(t *testing.T) {
		_, err := repo.GetAssignableRoleByPublicID(tenant.NewRoot(), "missing-role")
		if !errors.Is(err, core.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRepositoryAssignRole(t *testing.T) {
	t.Run("updates role and returns response", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		role, user := seedDB(t, db)

		// Create a second role to assign to
		newRole := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-editor"}, Name: "Editor", Slug: "editor"}
		if err := db.Create(&newRole).Error; err != nil {
			t.Fatalf("seed new role: %v", err)
		}

		fc := &fakeCache{}
		repo := NewRepository(db, fc)

		resp, err := repo.AssignRole(tenant.NewRoot(), &user, newRole.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.RoleID != newRole.ID {
			t.Fatalf("expected role_id %d, got %d", newRole.ID, resp.RoleID)
		}
		if resp.RoleName != "Editor" {
			t.Fatalf("expected role name Editor, got %s", resp.RoleName)
		}

		// Verify cache was invalidated
		expectedKey := "rbac:permissions:user-1"
		if len(fc.deletedKeys) != 1 || fc.deletedKeys[0] != expectedKey {
			t.Fatalf("expected cache delete for %s, got %v", expectedKey, fc.deletedKeys)
		}

		// Verify DB was updated
		var updated core.IAMUser
		if err := db.Where("public_id = ?", "user-1").First(&updated).Error; err != nil {
			t.Fatalf("fetch updated user: %v", err)
		}
		if updated.RoleID != newRole.ID {
			t.Fatalf("expected DB role_id %d, got %d", newRole.ID, updated.RoleID)
		}

		// Suppress unused variable warning
		_ = role
	})

	t.Run("works with nil cache", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		_, user := seedDB(t, db)

		repo := NewRepository(db, nil)

		resp, err := repo.AssignRole(tenant.NewRoot(), &user, user.RoleID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PublicID != "user-1" {
			t.Fatalf("expected user-1, got %s", resp.PublicID)
		}
	})
}

// NOTE: This repository delegates user lookup to queries.GetUserByPublicID and
// role lookup to queries.GetRoleByPublicID / queries.GetRoleBySlug.
// Full query behavior coverage lives in internal/modules/iam/queries/*_test.go.
