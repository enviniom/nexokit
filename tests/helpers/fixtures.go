package helpers

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/tests/fixtures"
	"gorm.io/gorm"
)

func SeedRole(t *testing.T, db *gorm.DB, slug string) roles.Role {
	t.Helper()
	role := fixtures.BuildRole(slug)
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

func SeedCompany(t *testing.T, db *gorm.DB, slug string) companies.Company {
	t.Helper()
	company := fixtures.BuildCompany(slug)
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	return company
}

func SeedPermission(t *testing.T, db *gorm.DB, module, action string) permissions.Permission {
	t.Helper()
	permission := fixtures.BuildPermission(module, action)
	if err := db.Create(&permission).Error; err != nil {
		t.Fatalf("seed permission: %v", err)
	}
	return permission
}

func SeedUserWithRole(t *testing.T, db *gorm.DB, publicID string, role roles.Role, company *companies.Company) users.User {
	t.Helper()
	var companyID *uint
	if company != nil {
		companyID = &company.ID
	}
	user := fixtures.BuildUser(publicID, role.ID, companyID)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}
