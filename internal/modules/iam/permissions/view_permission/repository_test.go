package view_permission

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

	// NOTE: this repository is a thin wrapper around queries.GetPermissionByPublicID.
	// Full query behavior coverage lives in internal/modules/iam/queries/get_permission_by_public_id_test.go.
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
