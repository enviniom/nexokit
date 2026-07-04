package core

import "testing"

func TestOnboardingCompany_TableName(t *testing.T) {
	if got, want := (OnboardingCompany{}).TableName(), "companies"; got != want {
		t.Errorf("OnboardingCompany.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingCompanyDomain_TableName(t *testing.T) {
	if got, want := (OnboardingCompanyDomain{}).TableName(), "company_domains"; got != want {
		t.Errorf("OnboardingCompanyDomain.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingUser_TableName(t *testing.T) {
	if got, want := (OnboardingUser{}).TableName(), "users"; got != want {
		t.Errorf("OnboardingUser.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingRole_TableName(t *testing.T) {
	if got, want := (OnboardingRole{}).TableName(), "roles"; got != want {
		t.Errorf("OnboardingRole.TableName() = %q, want %q", got, want)
	}
}

func TestOnboardingPermission_TableName(t *testing.T) {
	if got, want := (OnboardingPermission{}).TableName(), "permissions"; got != want {
		t.Errorf("OnboardingPermission.TableName() = %q, want %q", got, want)
	}
}
