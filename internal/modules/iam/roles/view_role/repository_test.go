package view_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryGetByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMPermission{}, &core.IAMRole{}, &core.IAMRolePermission{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	company := core.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: "company-1"}, Name: "Acme", Slug: "acme"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}

	permA := core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-a"}, Slug: "roles.view", Name: "View", Module: "roles", Action: "view"}
	permB := core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-b"}, Slug: "roles.edit", Name: "Edit", Module: "roles", Action: "edit"}
	if err := db.Create(&permA).Error; err != nil {
		t.Fatalf("seed permission A: %v", err)
	}
	if err := db.Create(&permB).Error; err != nil {
		t.Fatalf("seed permission B: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin", Slug: "admin", CompanyID: &company.ID}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	if err := db.Model(&role).Association("Permissions").Append(&permB, &permA); err != nil {
		t.Fatalf("seed role permissions: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: this repository is a thin wrapper around queries.GetRoleByPublicID.
	// Full query behavior coverage lives in internal/modules/iam/queries/get_role_by_public_id_test.go.
	item, err := repo.GetByPublicID(tenant.NewRoot(), "role-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID != "role-1" {
		t.Fatalf("expected public id role-1, got %s", item.PublicID)
	}
	if len(item.Permissions) != 2 || item.Permissions[0] != "roles.edit" || item.Permissions[1] != "roles.view" {
		t.Fatalf("expected sorted permissions [roles.edit roles.view], got %v", item.Permissions)
	}

	_, err = repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
