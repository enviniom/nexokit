package view_user

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
	if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	companyID := uint(42)
	user := core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		CompanyID:    &companyID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: this repository is a thin wrapper around queries.GetUserByPublicID.
	// Full query behavior coverage lives in internal/modules/iam/queries/get_user_by_public_id_test.go.
	item, err := repo.GetByPublicID(tenant.NewRoot(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID != "user-1" {
		t.Fatalf("expected public id user-1, got %s", item.PublicID)
	}
	if item.Name != "Alice" {
		t.Fatalf("expected name Alice, got %s", item.Name)
	}
	if item.RoleName != "Admin" {
		t.Fatalf("expected role name Admin, got %s", item.RoleName)
	}
	if item.CompanyID == nil || *item.CompanyID != 42 {
		t.Fatalf("expected company_id 42, got %v", item.CompanyID)
	}

	_, err = repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
