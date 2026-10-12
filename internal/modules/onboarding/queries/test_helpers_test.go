package queries

import (
	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	company_domainentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company_domain"
	permissionentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/permission"
	userentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/user"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(
		&companyentity.OnboardingCompany{},
		&company_domainentity.OnboardingCompanyDomain{},
		&userentity.OnboardingUser{},
		&permissionentity.OnboardingPermission{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := db.Exec("CREATE TABLE role_permissions (role_id integer not null, permission_id integer not null)").Error; err != nil {
		t.Fatalf("create role_permissions: %v", err)
	}

	return db
}
