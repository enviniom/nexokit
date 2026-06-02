package update_permission

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryGetPermissionByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMPermission{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	item := core.IAMPermission{
		BaseModel: shared.BaseModel{PublicID: "perm-1"},
		Slug:      "permissions.manage",
		Name:      "Manage",
		Module:    "permissions",
		Action:    "manage",
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed permission: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: this wrapper delegates the query details to queries.GetPermissionByPublicID.
	// Full query behavior coverage belongs to internal/modules/iam/queries tests.
	permission, err := repo.GetPermissionByPublicID("perm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if permission.PublicID != "perm-1" {
		t.Fatalf("expected public id perm-1, got %s", permission.PublicID)
	}

	_, err = repo.GetPermissionByPublicID("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGormRepositoryUpdatePermission(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMPermission{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	first := core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-1"}, Slug: "permissions.manage", Name: "Manage", Module: "permissions", Action: "manage"}
	second := core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-2"}, Slug: "permissions.view", Name: "View", Module: "permissions", Action: "view"}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("seed first: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("seed second: %v", err)
	}

	repo := NewRepository(db)

	first.Name = "Manage updated"
	first.Description = "updated"
	if err := repo.UpdatePermission(&first); err != nil {
		t.Fatalf("unexpected error updating permission: %v", err)
	}

	var stored core.IAMPermission
	if err := db.Where("public_id = ?", "perm-1").First(&stored).Error; err != nil {
		t.Fatalf("read updated permission: %v", err)
	}
	if stored.Name != "Manage updated" || stored.Description != "updated" {
		t.Fatalf("unexpected persisted values: %#v", stored)
	}

	first.Slug = second.Slug
	err = repo.UpdatePermission(&first)
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict for unique slug collision, got %v", err)
	}
}
