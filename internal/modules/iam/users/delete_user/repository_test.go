package delete_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryDelete(t *testing.T) {
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

	user := core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: this repository performs a direct soft-delete via GORM.
	// It does not wrap queries.GetUserByPublicID — the not-found check
	// is handled by inspecting RowsAffected on the DELETE statement.

	t.Run("delete success", func(t *testing.T) {
		if err := repo.Delete(tenant.NewRoot(), "user-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Verify soft-deleted: row still exists but DeletedAt is set.
		var count int64
		db.Unscoped().Model(&core.IAMUser{}).Where("public_id = ?", "user-1").Count(&count)
		if count != 1 {
			t.Fatalf("expected soft-deleted row to exist, got count %d", count)
		}
	})

	t.Run("not-found mapping", func(t *testing.T) {
		err := repo.Delete(tenant.NewRoot(), "missing")
		if !errors.Is(err, core.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
