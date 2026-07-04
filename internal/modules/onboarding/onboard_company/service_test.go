package onboard_company

import (
	"context"
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakePasswordHasher struct{}

func (f fakePasswordHasher) HashPassword(password string) (string, error) {
	return "hashed_" + password, nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&core.OnboardingCompany{}, &core.OnboardingCompanyDomain{}, &core.OnboardingRole{}, &core.OnboardingPermission{}, &core.OnboardingUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("CREATE TABLE role_permissions (role_id integer not null, permission_id integer not null)").Error; err != nil {
		t.Fatalf("create role_permissions: %v", err)
	}
	return db
}

func seedPermissions(t *testing.T, db *gorm.DB) {
	t.Helper()
	perms := []core.OnboardingPermission{
		{Slug: "users.create", Name: "Create Users", Module: "users", Action: "create", IsSystem: true},
		{Slug: "users.view", Name: "View Users", Module: "users", Action: "view", IsSystem: true},
		{Slug: "roles.view", Name: "View Roles", Module: "roles", Action: "view", IsSystem: true},
	}
	ids := []string{"perm_users_create", "perm_users_view", "perm_roles_view"}
	for i := range perms {
		perms[i].PublicID = ids[i]
		if err := db.Create(&perms[i]).Error; err != nil {
			t.Fatalf("seed permission: %v", err)
		}
	}
}

func newService(db *gorm.DB, platformDomain string) Service {
	return NewService(NewRepository(db), fakePasswordHasher{}, platformDomain)
}

func baseRequest() core.OnboardCompanyRequest {
	return core.OnboardCompanyRequest{Name: "Acme Inc", Slug: "acme", AdminName: "Jane Admin", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
}

func TestService_Onboard_Success(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	res, err := svc.Onboard(context.Background(), baseRequest())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if res.CompanySlug != "acme" || res.AdminEmail != "jane@acme.com" {
		t.Fatalf("unexpected response: %#v", res)
	}
}

func TestService_Onboard_CreatesPrimaryDomain(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	domain := " DreamMakers.COM.CO "
	req := baseRequest()
	req.Slug = "dreammakers"
	req.AdminEmail = "jane@dreammakers.com.co"
	req.Domain = &domain
	if _, err := svc.Onboard(context.Background(), req); err != nil {
		t.Fatalf("onboard: %v", err)
	}
	var d core.OnboardingCompanyDomain
	if err := db.Where("domain = ?", "dreammakers.com.co").First(&d).Error; err != nil {
		t.Fatalf("expected primary domain: %v", err)
	}
	if d.Kind != core.DomainKindPrimary {
		t.Fatalf("unexpected kind: %s", d.Kind)
	}
}

func TestService_Onboard_CreatesTechnicalDomainWhenRequested(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "KilaShop.COM")
	req := baseRequest()
	req.Slug = " DreamMakers "
	req.AdminEmail = "jane@dreammakers.com.co"
	req.GenerateTechnicalDomain = true
	if _, err := svc.Onboard(context.Background(), req); err != nil {
		t.Fatalf("onboard: %v", err)
	}
	var d core.OnboardingCompanyDomain
	if err := db.Where("domain = ?", "dreammakers.kilashop.com").First(&d).Error; err != nil {
		t.Fatalf("expected technical domain: %v", err)
	}
}

func TestService_Onboard_SkipsTechnicalDomainWhenNotRequested(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "kilashop.com")
	req := baseRequest()
	req.Slug = "dreammakers"
	req.AdminEmail = "jane@dreammakers.com.co"
	if _, err := svc.Onboard(context.Background(), req); err != nil {
		t.Fatalf("onboard: %v", err)
	}
	var count int64
	db.Model(&core.OnboardingCompanyDomain{}).Where("domain = ?", "dreammakers.kilashop.com").Count(&count)
	if count != 0 {
		t.Fatalf("expected technical domain skipped, got %d", count)
	}
}

func TestService_Onboard_DuplicateDomain_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	_ = db.Create(&core.OnboardingCompany{Slug: "existing", Name: "Existing", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&core.OnboardingCompanyDomain{CompanyID: 1, Domain: "dreammakers.com.co", Kind: core.DomainKindPrimary, Status: core.DomainStatusActive}).Error
	domain := "dreammakers.com.co"
	req := baseRequest()
	req.Slug = "dreammakers"
	req.AdminEmail = "jane@dreammakers.com.co"
	req.Domain = &domain
	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected duplicate domain, got: %v", err)
	}
}

func TestService_Onboard_DuplicateDomainWithTrailingDot_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	_ = db.Create(&core.OnboardingCompany{Slug: "existing", Name: "Existing", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&core.OnboardingCompanyDomain{CompanyID: 1, Domain: "dreammakers.com.co", Kind: core.DomainKindPrimary, Status: core.DomainStatusActive}).Error
	domain := "dreammakers.com.co."
	req := baseRequest()
	req.Slug = "dreammakers"
	req.AdminEmail = "jane@dreammakers.com.co"
	req.Domain = &domain
	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, core.ErrDuplicateCompanyDomain) {
		t.Fatalf("expected duplicate domain, got: %v", err)
	}
}

func TestService_Onboard_DuplicateTechnicalDomain_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "kilashop.com")
	_ = db.Create(&core.OnboardingCompany{Slug: "existing", Name: "Existing", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&core.OnboardingCompanyDomain{CompanyID: 1, Domain: "dreammakers.kilashop.com", Kind: core.DomainKindTechnical, Status: core.DomainStatusActive}).Error
	req := baseRequest()
	req.Slug = "dreammakers"
	req.AdminEmail = "jane@dreammakers.com.co"
	req.GenerateTechnicalDomain = true
	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, core.ErrDuplicateTechnicalDomain) {
		t.Fatalf("expected duplicate technical domain, got: %v", err)
	}
}

func TestService_Onboard_PrimaryDomainCannotEqualGeneratedTechnicalDomain(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "kilashop.com")
	domain := "dreammakers.kilashop.com"
	req := baseRequest()
	req.Slug = "dreammakers"
	req.AdminEmail = "jane@dreammakers.com.co"
	req.Domain = &domain
	req.GenerateTechnicalDomain = true
	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, core.ErrDuplicateTechnicalDomain) {
		t.Fatalf("expected duplicate technical domain, got: %v", err)
	}
}

func TestService_Onboard_DuplicateSlug_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	_ = db.Create(&core.OnboardingCompany{Name: "Acme Existing", Slug: "acme", Status: core.CompanyStatusActive}).Error
	_, err := svc.Onboard(context.Background(), baseRequest())
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected duplicate slug, got: %v", err)
	}
}

func TestService_Onboard_DuplicateEmail_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	_ = db.Create(&core.OnboardingUser{Name: "Jane Existing", Email: "jane@acme.com", PasswordHash: "secret", RoleID: 1}).Error
	req := baseRequest()
	req.Slug = "acme-new"
	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, core.ErrDuplicateAdminEmail) {
		t.Fatalf("expected duplicate email, got: %v", err)
	}
}
