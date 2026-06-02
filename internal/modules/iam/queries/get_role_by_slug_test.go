package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetRoleBySlug(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role := core.IAMRole{
		BaseModel: shared.BaseModel{PublicID: "role-root"},
		Name:      "Root",
		Slug:      "root",
		IsSystem:  true,
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	t.Run("returns role when slug exists", func(t *testing.T) {
		got, err := GetRoleBySlug(db, "root")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Slug != "root" {
			t.Fatalf("expected slug root, got %s", got.Slug)
		}
		if got.Name != "Root" {
			t.Fatalf("expected name Root, got %s", got.Name)
		}
	})

	t.Run("returns gorm.ErrRecordNotFound for missing slug", func(t *testing.T) {
		_, err := GetRoleBySlug(db, "nonexistent")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
		}
	})
}
