package toggle_user_status

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func seedToggleDB(t *testing.T, db *gorm.DB) {
	t.Helper()
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
}

func TestGormRepositoryToggleStatus(t *testing.T) {
	tests := []struct {
		name      string
		publicID  string
		isActive  bool
		seed      func(t *testing.T, db *gorm.DB)
		wantErr   error
		assertFn  func(t *testing.T, resp *core.UserResponse)
	}{
		{
			name:     "deactivates active user",
			publicID: "user-1",
			isActive: false,
			seed:     seedToggleDB,
			assertFn: func(t *testing.T, resp *core.UserResponse) {
				t.Helper()
				if resp.IsActive {
					t.Fatalf("expected IsActive false, got true")
				}
				if resp.PublicID != "user-1" {
					t.Fatalf("expected public id user-1, got %s", resp.PublicID)
				}
				if resp.RoleName != "Admin" {
					t.Fatalf("expected role name Admin, got %s", resp.RoleName)
				}
			},
		},
		{
			name:     "activates inactive user",
			publicID: "user-1",
			isActive: true,
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				seedToggleDB(t, db)
				db.Model(&core.IAMUser{}).Where("public_id = ?", "user-1").Update("is_active", false)
			},
			assertFn: func(t *testing.T, resp *core.UserResponse) {
				t.Helper()
				if !resp.IsActive {
					t.Fatalf("expected IsActive true, got false")
				}
			},
		},
		{
			name:     "returns ErrNotFound for missing user",
			publicID: "missing",
			isActive: false,
			seed:     seedToggleDB,
			wantErr:  core.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("open db: %v", err)
			}
			tt.seed(t, db)

			repo := NewRepository(db)
			resp, err := repo.ToggleStatus(tenant.NewRoot(), tt.publicID, tt.isActive)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.assertFn != nil {
				tt.assertFn(t, resp)
			}
		})
	}
}

// NOTE: This repository delegates user lookup to queries.GetUserByPublicID.
// Full query behavior coverage lives in internal/modules/iam/queries/get_user_by_public_id_test.go.
