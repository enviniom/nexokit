package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakePasswordHasher struct{}

func (f fakePasswordHasher) HashPassword(password string) (string, error) {
	return "hashed_" + password, nil
}

func (f fakePasswordHasher) VerifyPassword(password, hash string) error {
	if hash == "hashed_"+password {
		return nil
	}
	return errors.New("password mismatch")
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test DB: %v", err)
	}

	// AutoMigrate all models to set up the schema
	err = db.AutoMigrate(
		&companies.Company{},
		&companies.CompanyDomain{},
		&roles.Role{},
		&permissions.Permission{},
		&users.User{},
	)
	if err != nil {
		t.Fatalf("failed to auto migrate schema: %v", err)
	}

	return db
}

func TestService_Onboard_Success(t *testing.T) {
	db := setupTestDB(t)

	// Seed system permissions
	perms := []permissions.Permission{
		{BaseModel: shared.BaseModel{PublicID: "perm-users-create"}, Slug: "users.create", Name: "Create Users", Module: "users", Action: "create"},
		{BaseModel: shared.BaseModel{PublicID: "perm-users-view"}, Slug: "users.view", Name: "View Users", Module: "users", Action: "view"},
		{BaseModel: shared.BaseModel{PublicID: "perm-roles-view"}, Slug: "roles.view", Name: "View Roles", Module: "roles", Action: "view"},
	}
	for i := range perms {
		if err := db.Create(&perms[i]).Error; err != nil {
			t.Fatalf("failed to seed permission: %v", err)
		}
	}

	svc := NewService(db, fakePasswordHasher{})

	req := OnboardCompanyRequest{
		Name:          "Acme Inc",
		Slug:          "acme",
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@acme.com",
		AdminPassword: "SuperSecurePassword123",
	}

	res, err := svc.Onboard(context.Background(), req)
	if err != nil {
		t.Fatalf("expected successful onboarding, got error: %v", err)
	}

	if res.CompanySlug != "acme" {
		t.Errorf("expected company slug 'acme', got: %s", res.CompanySlug)
	}

	// Verify Company was created
	var company companies.Company
	if err := db.Where("slug = ?", "acme").First(&company).Error; err != nil {
		t.Errorf("expected company to be created in DB: %v", err)
	}

	// Verify Tenant Roles were created
	var adminRole roles.Role
	if err := db.Where("company_id = ? AND slug = ?", company.ID, roles.AdminRoleSlug).Preload("Permissions").First(&adminRole).Error; err != nil {
		t.Errorf("expected admin role to be created: %v", err)
	}

	// Verify admin role has all 3 permissions mapped
	if len(adminRole.Permissions) != 3 {
		t.Errorf("expected admin role to have 3 permissions, got %d", len(adminRole.Permissions))
	}

	var userRole roles.Role
	if err := db.Where("company_id = ? AND slug = ?", company.ID, roles.UserRoleSlug).Preload("Permissions").First(&userRole).Error; err != nil {
		t.Errorf("expected user role to be created: %v", err)
	}

	// Verify user role only has users.view and roles.view
	if len(userRole.Permissions) != 2 {
		t.Errorf("expected user role to have 2 permissions, got %d", len(userRole.Permissions))
	}

	// Verify Admin User was created with correct password hash and role
	var adminUser users.User
	if err := db.Where("company_id = ? AND email = ?", company.ID, "jane@acme.com").First(&adminUser).Error; err != nil {
		t.Errorf("expected admin user to be created: %v", err)
	}

	if adminUser.RoleID != adminRole.ID {
		t.Errorf("expected admin user to have admin role ID %d, got %d", adminRole.ID, adminUser.RoleID)
	}

	if adminUser.PasswordHash != "hashed_SuperSecurePassword123" {
		t.Errorf("expected password to be hashed correctly")
	}
}

func TestService_Onboard_CreatesPrimaryDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, fakePasswordHasher{})
	domain := " DreamMakers.COM.CO "

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:          "Dream Makers",
		Slug:          "dreammakers",
		Domain:        &domain,
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@dreammakers.com.co",
		AdminPassword: "SuperSecurePassword123",
	})
	if err != nil {
		t.Fatalf("expected successful onboarding, got error: %v", err)
	}

	var company companies.Company
	if err := db.Where("slug = ?", "dreammakers").First(&company).Error; err != nil {
		t.Fatalf("expected company: %v", err)
	}
	var domainRow companies.CompanyDomain
	if err := db.Where("company_id = ? AND domain = ?", company.ID, "dreammakers.com.co").First(&domainRow).Error; err != nil {
		t.Fatalf("expected primary domain row: %v", err)
	}
	if domainRow.Kind != companies.CompanyDomainKindPrimary || domainRow.Status != companies.CompanyDomainStatusActive || domainRow.RedirectToPrimary {
		t.Fatalf("unexpected primary domain row: %#v", domainRow)
	}
}

func TestService_Onboard_CreatesTechnicalDomainWhenRequested(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, fakePasswordHasher{}, WithPlatformDomain("KilaShop.COM"))

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:                    "Dream Makers",
		Slug:                    " DreamMakers ",
		GenerateTechnicalDomain: true,
		AdminName:               "Jane Admin",
		AdminEmail:              "jane@dreammakers.com.co",
		AdminPassword:           "SuperSecurePassword123",
	})
	if err != nil {
		t.Fatalf("expected successful onboarding, got error: %v", err)
	}

	var domainRow companies.CompanyDomain
	if err := db.Where("domain = ?", "dreammakers.kilashop.com").First(&domainRow).Error; err != nil {
		t.Fatalf("expected technical domain row: %v", err)
	}
	if domainRow.Kind != companies.CompanyDomainKindTechnical || domainRow.Status != companies.CompanyDomainStatusActive {
		t.Fatalf("unexpected technical domain row: %#v", domainRow)
	}
}

func TestService_Onboard_SkipsTechnicalDomainWhenNotRequested(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, fakePasswordHasher{}, WithPlatformDomain("kilashop.com"))

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:          "Dream Makers",
		Slug:          "dreammakers",
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@dreammakers.com.co",
		AdminPassword: "SuperSecurePassword123",
	})
	if err != nil {
		t.Fatalf("expected successful onboarding, got error: %v", err)
	}

	var domainCount int64
	db.Model(&companies.CompanyDomain{}).Where("domain = ?", "dreammakers.kilashop.com").Count(&domainCount)
	if domainCount != 0 {
		t.Fatalf("expected technical domain to be skipped, got count %d", domainCount)
	}
}

func TestService_Onboard_DuplicateDomain_Rollback(t *testing.T) {
	db := setupTestDB(t)
	existingCompany := companies.Company{Name: "Existing", Slug: "existing", Status: companies.CompanyStatusActive}
	if err := db.Create(&existingCompany).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	if err := db.Create(&companies.CompanyDomain{CompanyID: existingCompany.ID, Domain: "dreammakers.com.co", Kind: companies.CompanyDomainKindPrimary, Status: companies.CompanyDomainStatusActive}).Error; err != nil {
		t.Fatalf("seed company domain: %v", err)
	}
	svc := NewService(db, fakePasswordHasher{})
	domain := "dreammakers.com.co"

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:          "Dream Makers",
		Slug:          "dreammakers",
		Domain:        &domain,
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@dreammakers.com.co",
		AdminPassword: "SuperSecurePassword123",
	})
	if !errors.Is(err, ErrDuplicateCompanyDomain) {
		t.Fatalf("expected ErrDuplicateCompanyDomain, got: %v", err)
	}

	var companyCount int64
	db.Model(&companies.Company{}).Where("slug = ?", "dreammakers").Count(&companyCount)
	if companyCount > 0 {
		t.Fatal("expected duplicate domain onboarding to rollback company creation")
	}
}

func TestService_Onboard_DuplicateDomainWithTrailingDot_Rollback(t *testing.T) {
	db := setupTestDB(t)
	existingCompany := companies.Company{Name: "Existing", Slug: "existing", Status: companies.CompanyStatusActive}
	if err := db.Create(&existingCompany).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	if err := db.Create(&companies.CompanyDomain{CompanyID: existingCompany.ID, Domain: "dreammakers.com.co", Kind: companies.CompanyDomainKindPrimary, Status: companies.CompanyDomainStatusActive}).Error; err != nil {
		t.Fatalf("seed company domain: %v", err)
	}
	svc := NewService(db, fakePasswordHasher{})
	domain := "dreammakers.com.co."

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:          "Dream Makers",
		Slug:          "dreammakers",
		Domain:        &domain,
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@dreammakers.com.co",
		AdminPassword: "SuperSecurePassword123",
	})
	if !errors.Is(err, ErrDuplicateCompanyDomain) {
		t.Fatalf("expected ErrDuplicateCompanyDomain, got: %v", err)
	}

	var companyCount int64
	db.Model(&companies.Company{}).Where("slug = ?", "dreammakers").Count(&companyCount)
	if companyCount > 0 {
		t.Fatal("expected duplicate domain onboarding to rollback company creation")
	}
}

func TestService_Onboard_DuplicateTechnicalDomain_Rollback(t *testing.T) {
	db := setupTestDB(t)
	existingCompany := companies.Company{Name: "Existing", Slug: "existing", Status: companies.CompanyStatusActive}
	if err := db.Create(&existingCompany).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}
	if err := db.Create(&companies.CompanyDomain{CompanyID: existingCompany.ID, Domain: "dreammakers.kilashop.com", Kind: companies.CompanyDomainKindTechnical, Status: companies.CompanyDomainStatusActive}).Error; err != nil {
		t.Fatalf("seed technical domain: %v", err)
	}
	svc := NewService(db, fakePasswordHasher{}, WithPlatformDomain("kilashop.com"))

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:                    "Dream Makers",
		Slug:                    "dreammakers",
		GenerateTechnicalDomain: true,
		AdminName:               "Jane Admin",
		AdminEmail:              "jane@dreammakers.com.co",
		AdminPassword:           "SuperSecurePassword123",
	})
	if !errors.Is(err, ErrDuplicateTechnicalDomain) {
		t.Fatalf("expected ErrDuplicateTechnicalDomain, got: %v", err)
	}

	var companyCount int64
	db.Model(&companies.Company{}).Where("slug = ?", "dreammakers").Count(&companyCount)
	if companyCount > 0 {
		t.Fatal("expected duplicate technical domain onboarding to rollback company creation")
	}
}

func TestService_Onboard_PrimaryDomainCannotEqualGeneratedTechnicalDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, fakePasswordHasher{}, WithPlatformDomain("kilashop.com"))
	domain := "dreammakers.kilashop.com"

	_, err := svc.Onboard(context.Background(), OnboardCompanyRequest{
		Name:                    "Dream Makers",
		Slug:                    "dreammakers",
		Domain:                  &domain,
		GenerateTechnicalDomain: true,
		AdminName:               "Jane Admin",
		AdminEmail:              "jane@dreammakers.com.co",
		AdminPassword:           "SuperSecurePassword123",
	})
	if !errors.Is(err, ErrDuplicateTechnicalDomain) {
		t.Fatalf("expected ErrDuplicateTechnicalDomain, got: %v", err)
	}

	var companyCount int64
	db.Model(&companies.Company{}).Where("slug = ?", "dreammakers").Count(&companyCount)
	if companyCount > 0 {
		t.Fatal("expected matching primary/technical domain onboarding to rollback company creation")
	}
}

func TestService_Onboard_DuplicateSlug_Rollback(t *testing.T) {
	db := setupTestDB(t)

	// Pre-create a company with slug "acme"
	preCompany := companies.Company{
		Name:   "Acme Existing",
		Slug:   "acme",
		Status: companies.CompanyStatusActive,
	}
	db.Create(&preCompany)

	svc := NewService(db, fakePasswordHasher{})

	req := OnboardCompanyRequest{
		Name:          "Acme Inc New",
		Slug:          "acme", // duplicate slug
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@acme.com",
		AdminPassword: "SuperSecurePassword123",
	}

	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, ErrDuplicateCompanySlug) {
		t.Fatalf("expected ErrDuplicateCompanySlug, got: %v", err)
	}

	// Verify no user "jane@acme.com" was created due to transaction rollback
	var userCount int64
	db.Model(&users.User{}).Where("email = ?", "jane@acme.com").Count(&userCount)
	if userCount > 0 {
		t.Errorf("expected admin user creation to rollback, but user exists")
	}
}

func TestService_Onboard_DuplicateEmail_Rollback(t *testing.T) {
	db := setupTestDB(t)

	// Pre-create a user with email "jane@acme.com"
	preUser := users.User{
		Name:         "Jane Existing",
		Email:        "jane@acme.com",
		PasswordHash: "secret",
		RoleID:       1,
	}
	db.Create(&preUser)

	svc := NewService(db, fakePasswordHasher{})

	req := OnboardCompanyRequest{
		Name:          "Acme Inc New",
		Slug:          "acme-new",
		AdminName:     "Jane Admin",
		AdminEmail:    "jane@acme.com", // duplicate email
		AdminPassword: "SuperSecurePassword123",
	}

	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, ErrDuplicateAdminEmail) {
		t.Fatalf("expected ErrDuplicateAdminEmail, got: %v", err)
	}

	// Verify company "acme-new" was rolled back and does not exist in DB
	var companyCount int64
	db.Model(&companies.Company{}).Where("slug = ?", "acme-new").Count(&companyCount)
	if companyCount > 0 {
		t.Errorf("expected company creation to rollback, but company exists")
	}
}
