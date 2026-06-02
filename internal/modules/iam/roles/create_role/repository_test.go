package create_role

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	company := core.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: "comp-1"}, Name: "Acme", Slug: "acme"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}

	repo := NewRepository(db)
	item, err := repo.Create(tenant.NewScoped(company.ID, company.Slug), core.CreateRoleRequest{Name: "Manager", Slug: "manager", Description: "desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID == "" {
		t.Fatalf("expected generated public id")
	}
	if item.CompanyID == nil || *item.CompanyID != company.ID {
		t.Fatalf("expected company id %d, got %#v", company.ID, item.CompanyID)
	}
}

func TestGormRepositoryQueryWrappers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	if err := db.Create(&core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Manager", Slug: "manager"}).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: wrappers delegate matching behavior to reusable query functions.
	// Full query behavior coverage belongs to internal/modules/iam/queries tests.
	byName, err := repo.ExistsRoleByName(tenant.NewRoot(), "Manager")
	if err != nil {
		t.Fatalf("exists by name: %v", err)
	}
	if !byName {
		t.Fatalf("expected role name to exist")
	}

	bySlug, err := repo.ExistsRoleBySlug(tenant.NewRoot(), "manager")
	if err != nil {
		t.Fatalf("exists by slug: %v", err)
	}
	if !bySlug {
		t.Fatalf("expected role slug to exist")
	}
}
