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
	repo := NewRepository(db)
	if _, err := repo.CreateCompany("Existing", "acme"); err != nil {
		t.Fatalf("seed company: %v", err)
	}
	err := repo.EnsureCompanySlugAvailable("acme")
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected duplicate slug error, got: %v", err)
	}
}

func TestRepository_CreateAndAssignRecords(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository(db)

	company, err := repo.CreateCompany("Acme", "acme")
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	if err := repo.CreateCompanyDomain(company.ID, "acme.com", core.DomainKindPrimary); err != nil {
		t.Fatalf("create domain: %v", err)
	}

	role, err := repo.CreateRole(company.ID, "Admin", core.RoleSlugAdmin, "admin role")
	if err != nil {
		t.Fatalf("create role: %v", err)
	}

	perm := core.OnboardingPermission{Slug: "users.view", Name: "View", Module: "users", Action: "view", IsSystem: true}
	perm.PublicID = "perm_users_view"
	if err := db.Create(&perm).Error; err != nil {
		t.Fatalf("seed permission: %v", err)
	}
	if err := repo.AssignPermissionToRole(role.ID, perm.ID); err != nil {
		t.Fatalf("assign permission: %v", err)
	}

	if _, err := repo.CreateAdminUser(company.ID, role.ID, "Jane", "jane@acme.com", "hashed"); err != nil {
		t.Fatalf("create admin user: %v", err)
	}
}

func TestRepository_CreateCompany_TranslatesUniqueViolation(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository(db)
	if _, err := repo.CreateCompany("Existing", "acme"); err != nil {
		t.Fatalf("seed company: %v", err)
	}
	_, err := repo.CreateCompany("Another", "acme")
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected duplicate slug sentinel, got: %v", err)
	}
}

func TestRepository_CreateCompanyDomain_TranslatesUniqueViolation(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository(db)
	if err := repo.CreateCompanyDomain(1, "acme.com", core.DomainKindPrimary); err != nil {
		t.Fatalf("seed domain: %v", err)
	}
	err := repo.CreateCompanyDomain(2, "acme.com", core.DomainKindPrimary)
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected duplicate company domain sentinel, got: %v", err)
	}
}

func TestRepository_CreateCompanyDomain_TranslatesTechnicalUniqueViolation(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository(db)
	if err := repo.CreateCompanyDomain(1, "acme.kilashop.com", core.DomainKindTechnical); err != nil {
		t.Fatalf("seed domain: %v", err)
	}
	err := repo.CreateCompanyDomain(2, "acme.kilashop.com", core.DomainKindTechnical)
	if !errors.Is(err, core.ErrDuplicateTechnicalDomain) {
		t.Fatalf("expected duplicate technical domain sentinel, got: %v", err)
	}
}

func TestRepository_CreateAdminUser_TranslatesUniqueViolation(t *testing.T) {
	db := repoTestDB(t)
	repo := NewRepository(db)
	if _, err := repo.CreateAdminUser(1, 1, "Jane", "jane@acme.com", "hashed"); err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	_, err := repo.CreateAdminUser(2, 1, "John", "jane@acme.com", "hashed")
	if !errors.Is(err, core.ErrDuplicateAdminEmail) {
		t.Fatalf("expected duplicate admin email sentinel, got: %v", err)
	}
}
