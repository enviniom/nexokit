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

func TestGetUserByPublicID(t *testing.T) {
	tests := []struct {
		name      string
		publicID  string
		seed      func(t *testing.T, db *gorm.DB)
		assertErr func(t *testing.T, err error)
	}{
		{
			name:     "returns user when found",
			publicID: "user-1",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
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
			publicID: "user-missing",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
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

			user, err := GetUserByPublicID(db, tenant.NewRoot(), tt.publicID)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			if user.PublicID != tt.publicID {
				t.Fatalf("expected public id %q, got %q", tt.publicID, user.PublicID)
			}
			if user.Role.Name != "Admin" {
				t.Fatalf("expected preloaded role name Admin, got %s", user.Role.Name)
			}
		})
	}
}
