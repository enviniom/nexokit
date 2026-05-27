package delete_company

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_DeleteSoftDeletes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.Company{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	company := core.Company{BaseModel: shared.BaseModel{PublicID: "company_delete"}, Name: "Delete", Slug: "delete", Status: core.CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := NewRepository(db)

	if err := repo.Delete("company_delete"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByPublicID("company_delete"); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
