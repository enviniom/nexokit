package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		seed      func(t *testing.T, db *gorm.DB)
		assertErr func(t *testing.T, err error)
		assertFn  func(t *testing.T, user *core.IAMUser)
	}{
		{
			name:  "returns user when found",
			email: "alice@example.com",
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
			assertFn: func(t *testing.T, user *core.IAMUser) {
				t.Helper()
				if user.Email != "alice@example.com" {
					t.Fatalf("expected email alice@example.com, got %s", user.Email)
				}
				if user.Name != "Alice" {
					t.Fatalf("expected name Alice, got %s", user.Name)
				}
			},
		},
		{
			name:  "returns gorm not found when missing",
			email: "nobody@example.com",
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

			user, err := GetUserByEmail(db, tt.email)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			if tt.assertFn != nil {
				tt.assertFn(t, user)
			}
		})
	}
}
