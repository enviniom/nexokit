package onboard_company

import (
	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	company_domainentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company_domain"
	permissionentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/permission"
	roleentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/role"
	userentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/user"
	"testing"
)

// Table names below match migrations/20260101000000_init.sql, not inferred AutoMigrate names.
func TestOnboardingCompany_TableName(t *testing.T) {
	if got, want := (companyentity.OnboardingCompany{}).TableName(), "companies"; got != want {
		t.Errorf("OnboardingCompany.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingCompanyDomain_TableName(t *testing.T) {
	if got, want := (company_domainentity.OnboardingCompanyDomain{}).TableName(), "company_domains"; got != want {
		t.Errorf("OnboardingCompanyDomain.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingUser_TableName(t *testing.T) {
	if got, want := (userentity.OnboardingUser{}).TableName(), "users"; got != want {
		t.Errorf("OnboardingUser.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingRole_TableName(t *testing.T) {
	if got, want := (roleentity.OnboardingRole{}).TableName(), "roles"; got != want {
		t.Errorf("OnboardingRole.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingPermission_TableName(t *testing.T) {
	if got, want := (permissionentity.OnboardingPermission{}).TableName(), "permissions"; got != want {
		t.Errorf("OnboardingPermission.TableName() = %q, want %q", got, want)
	}
}
