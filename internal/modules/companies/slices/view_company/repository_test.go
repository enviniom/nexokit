package view_company

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_GetByPublicID_WithDomains(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{})
	company := company.Company{BaseModel: shared.BaseModel{PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: core.CompanyStatusActive}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	if err := db.Create(&company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "dom_01"}, CompanyID: company.ID, Domain: "acme.com", Status: core.CompanyDomainStatusActive, Kind: core.CompanyDomainKindPrimary}).Error; err != nil {
		t.Fatalf("seed domain: %v", err)
	}

	repo := NewRepository(db)
	got, err := repo.GetByPublicID(tenant.NewRoot(), "cmp_01")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Domains) != 1 || got.Domains[0].PublicID != "dom_01" || got.Domains[0].CompanyID != company.ID {
		t.Fatalf("expected associated domain, got %+v", got.Domains)
	}
}

func TestGormRepository_TenantQueryBoundary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{}); err != nil {
		t.Fatal(err)
	}
	item := company.Company{BaseModel: shared.BaseModel{ID: 7, PublicID: "own"}, Name: "Own", Slug: "own"}
	if err = db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		tc   tenant.TenantContext
		want error
	}{
		{"own scope", tenant.NewScoped(7, "own"), nil},
		{"foreign scope", tenant.NewScoped(8, "foreign"), core.ErrCompanyNotFound},
		{"missing scope", tenant.TenantContext{}, core.ErrCompanyNotFound},
		{"root global", tenant.NewRoot(), nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRepository(db).GetByPublicID(tt.tc, "own")
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
			if tt.want != nil && got != nil {
				t.Fatal("foreign company returned")
			}
		})
	}
}

func TestGormRepository_GetByPublicID_NotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&company.Company{})

	repo := NewRepository(db)
	_, err := repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
