package update_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryGetRoleByPublicIDAndWrappers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	roleA := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-a"}, Name: "Manager", Slug: "manager"}
	roleB := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-b"}, Name: "Lead", Slug: "lead"}
	if err := db.Create(&roleA).Error; err != nil {
		t.Fatalf("seed role a: %v", err)
	}
	if err := db.Create(&roleB).Error; err != nil {
		t.Fatalf("seed role b: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: this wrapper delegates query details to queries.GetRoleByPublicID.
	// Full query behavior coverage belongs to internal/modules/iam/queries tests.
	got, err := repo.GetRoleByPublicID(tenant.NewRoot(), "role-a")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got.PublicID != "role-a" {
		t.Fatalf("expected role-a, got %s", got.PublicID)
	}

	_, err = repo.GetRoleByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// NOTE: wrappers delegate matching behavior to reusable query functions.
	// Full query behavior coverage belongs to internal/modules/iam/queries tests.
	nameConflict, err := repo.ExistsRoleByName(tenant.NewRoot(), "Manager", roleB.ID)
	if err != nil {
		t.Fatalf("exists by name: %v", err)
	}
	if !nameConflict {
		t.Fatalf("expected name conflict to be true")
	}

	nameSelfOnly, err := repo.ExistsRoleByName(tenant.NewRoot(), "Manager", roleA.ID)
	if err != nil {
		t.Fatalf("exists by name self: %v", err)
	}
	if nameSelfOnly {
		t.Fatalf("expected name conflict false when only self matches")
	}

	slugConflict, err := repo.ExistsRoleBySlug(tenant.NewRoot(), "manager", roleB.ID)
	if err != nil {
		t.Fatalf("exists by slug: %v", err)
	}
	if !slugConflict {
		t.Fatalf("expected slug conflict to be true")
	}

	slugSelfOnly, err := repo.ExistsRoleBySlug(tenant.NewRoot(), "manager", roleA.ID)
	if err != nil {
		t.Fatalf("exists by slug self: %v", err)
	}
	if slugSelfOnly {
		t.Fatalf("expected slug conflict false when only self matches")
	}
}

func TestGormRepositoryUpdateRole(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-a"}, Name: "Manager", Slug: "manager", Description: "desc"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	repo := NewRepository(db)
	role.Name = "Lead"
	role.Slug = "lead"
	role.Description = "updated"

	if err := repo.UpdateRole(tenant.NewRoot(), &role); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	var persisted core.IAMRole
	if err := db.Where("id = ?", role.ID).First(&persisted).Error; err != nil {
		t.Fatalf("read updated role: %v", err)
	}
	if persisted.Name != "Lead" || persisted.Slug != "lead" || persisted.Description != "updated" {
		t.Fatalf("unexpected persisted role state: %#v", persisted)
	}

	missing := core.IAMRole{BaseModel: shared.BaseModel{ID: 999}, Name: "x", Slug: "x"}
	if err := repo.UpdateRole(tenant.NewRoot(), &missing); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
