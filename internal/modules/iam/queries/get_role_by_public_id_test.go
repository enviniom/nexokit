package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetRoleByPublicID(t *testing.T) {
	tests := []struct {
		name      string
		publicID  string
		seed      func(t *testing.T, db *gorm.DB)
		assertErr func(t *testing.T, err error)
	}{
		{
			name:     "returns role when found",
			publicID: "role-1",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
				role := core.IAMRole{
					BaseModel: shared.BaseModel{PublicID: "role-1"},
					Name:      "Admin",
					Slug:      "admin",
				}
				if err := db.Create(&role).Error; err != nil {
					t.Fatalf("seed role: %v", err)
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
			publicID: "role-missing",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
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

			role, err := GetRoleByPublicID(db, tenant.NewRoot(), tt.publicID)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			if role.PublicID != tt.publicID {
				t.Fatalf("expected public id %q, got %q", tt.publicID, role.PublicID)
			}
		})
	}
}
