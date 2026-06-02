package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetPermissionByPublicID(t *testing.T) {
	tests := []struct {
		name      string
		publicID  string
		seed      func(t *testing.T, db *gorm.DB)
		assertErr func(t *testing.T, err error)
	}{
		{
			name:     "returns permission when found",
			publicID: "perm-1",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMPermission{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
				item := core.IAMPermission{
					BaseModel: shared.BaseModel{PublicID: "perm-1"},
					Slug:      "permissions.manage",
					Name:      "Manage permissions",
					Module:    "permissions",
					Action:    "manage",
				}
				if err := db.Create(&item).Error; err != nil {
					t.Fatalf("seed permission: %v", err)
				}
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name:     "returns gorm not found when missing",
			publicID: "perm-missing",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMPermission{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatalf("expected ErrRecordNotFound, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("open db: %v", err)
			}
			tt.seed(t, db)

			item, err := GetPermissionByPublicID(db, tt.publicID)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			if item.PublicID != tt.publicID {
				t.Fatalf("expected public id %q, got %q", tt.publicID, item.PublicID)
			}
		})
	}
}
