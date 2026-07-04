package list_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryListAllPermissions(t *testing.T) {
	tests := []struct {
		name      string
		seed      func(t *testing.T, db *gorm.DB)
		wantSlugs []string
		wantErr   bool
	}{
		{
			name: "returns ordered permissions by module display_order and slug",
			seed: func(t *testing.T, db *gorm.DB) {
				t.Helper()
				if err := db.AutoMigrate(&core.IAMPermission{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}

				items := []core.IAMPermission{
					{BaseModel: shared.BaseModel{PublicID: "3"}, Slug: "users.edit", Name: "Edit users", Module: "users", Action: "edit", DisplayOrder: 20},
					{BaseModel: shared.BaseModel{PublicID: "1"}, Slug: "roles.list", Name: "List roles", Module: "roles", Action: "list", DisplayOrder: 10},
					{BaseModel: shared.BaseModel{PublicID: "2"}, Slug: "roles.create", Name: "Create roles", Module: "roles", Action: "create", DisplayOrder: 10},
				}
				if err := db.Create(&items).Error; err != nil {
					t.Fatalf("seed permissions: %v", err)
				}
			},
			wantSlugs: []string{"roles.create", "roles.list", "users.edit"},
			wantErr:   false,
		},
		{
			name:    "returns error when table does not exist",
			seed:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("open db: %v", err)
			}

			if tt.seed != nil {
				tt.seed(t, db)
			}

			items, err := NewRepository(db).ListAllPermissions()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(items) != len(tt.wantSlugs) {
				t.Fatalf("expected %d permissions, got %d", len(tt.wantSlugs), len(items))
			}

			for i, slug := range tt.wantSlugs {
				if items[i].Slug != slug {
					t.Fatalf("unexpected order at %d: want %q got %q", i, slug, items[i].Slug)
				}
			}
		})
	}
}
