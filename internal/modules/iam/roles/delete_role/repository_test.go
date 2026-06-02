package delete_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryGetByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: this wrapper delegates query details to queries.GetRoleByPublicID.
	// Full query behavior coverage belongs to internal/modules/iam/queries tests.
	got, err := repo.GetByPublicID(tenant.NewRoot(), "role-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicID != "role-1" {
		t.Fatalf("expected role-1, got %s", got.PublicID)
	}

	_, err = repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGormRepositoryCountUsersByRoleID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	userA := core.IAMUser{BaseModel: shared.BaseModel{PublicID: "user-1"}, Name: "A", Email: "a@test.dev", PasswordHash: "hash", RoleID: role.ID, IsActive: true}
	userB := core.IAMUser{BaseModel: shared.BaseModel{PublicID: "user-2"}, Name: "B", Email: "b@test.dev", PasswordHash: "hash", RoleID: role.ID, IsActive: true}
	if err := db.Create(&userA).Error; err != nil {
		t.Fatalf("seed user a: %v", err)
	}
	if err := db.Create(&userB).Error; err != nil {
		t.Fatalf("seed user b: %v", err)
	}

	repo := NewRepository(db)
	count, err := repo.CountUsersByRoleID(role.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
}

func TestGormRepositoryDeleteByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	repo := NewRepository(db)
	if err := repo.DeleteByPublicID(tenant.NewRoot(), "role-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if err := repo.DeleteByPublicID(tenant.NewRoot(), "missing"); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
