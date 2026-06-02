package change_user_password

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func seedSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
}

func seedRole(t *testing.T, db *gorm.DB, publicID, name, slug string) core.IAMRole {
	t.Helper()
	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: publicID}, Name: name, Slug: slug}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

func seedUser(t *testing.T, db *gorm.DB, publicID, name, email, hash string, roleID uint) core.IAMUser {
	t.Helper()
	user := core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		RoleID:       roleID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func TestRepositoryGetByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")
	seedUser(t, db, "user-1", "Alice", "alice@example.com", "hash", role.ID)

	repo := NewRepository(db)

	// NOTE: GetByPublicID delegates to queries.GetUserByPublicID.
	// Full query behavior coverage lives in queries/get_user_by_public_id_test.go.
	user, err := repo.GetByPublicID(tenant.NewRoot(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.PublicID != "user-1" {
		t.Fatalf("expected public id user-1, got %s", user.PublicID)
	}
	if user.Role.Slug != "admin" {
		t.Fatalf("expected role slug admin, got %s", user.Role.Slug)
	}

	_, err = repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepositoryGetRoleBySlug(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	seedRole(t, db, "role-root", "Root", "root")

	repo := NewRepository(db)

	role, err := repo.GetRoleBySlug("root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Slug != "root" {
		t.Fatalf("expected slug root, got %s", role.Slug)
	}

	_, err = repo.GetRoleBySlug("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepositoryUpdatePassword(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")
	user := seedUser(t, db, "user-1", "Alice", "alice@example.com", "old-hash", role.ID)

	repo := NewRepository(db)

	if err := repo.UpdatePassword(user.ID, "new-hash"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var updated core.IAMUser
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updated.PasswordHash != "new-hash" {
		t.Fatalf("expected password hash new-hash, got %s", updated.PasswordHash)
	}
}

func TestRepositoryUpdatePasswordNonExistent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)

	repo := NewRepository(db)

	// GORM Update with no matching rows does not return an error by default.
	// This test documents that behavior — RowsAffected is 0 but no error.
	if err := repo.UpdatePassword(9999, "new-hash"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
