package onboard_company

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func repoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.OnboardingCompany{}, &core.OnboardingCompanyDomain{}, &core.OnboardingRole{}, &core.OnboardingPermission{}, &core.OnboardingUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("CREATE TABLE role_permissions (role_id integer not null, permission_id integer not null)").Error; err != nil {
		t.Fatalf("create role_permissions: %v", err)
	}
	return db
}

func TestRepository_DelegatesUniquenessChecks(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository()
	if _, err := repo.CreateCompany(db, "Existing", "acme"); err != nil {
		t.Fatalf("seed company: %v", err)
	}
	err := repo.EnsureCompanySlugAvailable(db, "acme")
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected duplicate slug error, got: %v", err)
	}
}

func TestRepository_CreateAndAssignRecords(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository()

	company, err := repo.CreateCompany(db, "Acme", "acme")
	if err != nil { t.Fatalf("create company: %v", err) }
	if err := repo.CreateCompanyDomain(db, company.ID, "acme.com", core.DomainKindPrimary); err != nil { t.Fatalf("create domain: %v", err) }

	role, err := repo.CreateRole(db, company.ID, "Admin", core.RoleSlugAdmin, "admin role")
	if err != nil { t.Fatalf("create role: %v", err) }

	perm := core.OnboardingPermission{Slug: "users.view", Name: "View", Module: "users", Action: "view", IsSystem: true}
	perm.PublicID = "perm_users_view"
	if err := db.Create(&perm).Error; err != nil { t.Fatalf("seed permission: %v", err) }
	if err := repo.AssignPermissionToRole(db, role.ID, perm.ID); err != nil { t.Fatalf("assign permission: %v", err) }

	if _, err := repo.CreateAdminUser(db, company.ID, role.ID, "Jane", "jane@acme.com", "hashed"); err != nil {
		t.Fatalf("create admin user: %v", err)
	}
}
