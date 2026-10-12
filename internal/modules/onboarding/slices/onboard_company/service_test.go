package onboard_company

import (
	"context"
	"errors"
	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	company_domainentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company_domain"
	permissionentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/permission"
	roleentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/role"
	userentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/user"
	"github.com/enviniom/nexokit/internal/shared"
	"testing"
	"time"

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
	if err := db.AutoMigrate(&companyentity.OnboardingCompany{}, &company_domainentity.OnboardingCompanyDomain{}, &roleentity.OnboardingRole{}, &permissionentity.OnboardingPermission{}, &userentity.OnboardingUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("CREATE TABLE role_permissions (role_id integer not null, permission_id integer not null)").Error; err != nil {
		t.Fatalf("create role_permissions: %v", err)
	}
	return db
}

func seedPermissions(t *testing.T, db *gorm.DB) {
	t.Helper()
	perms := []permissionentity.OnboardingPermission{
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

func baseRequest() companyentity.OnboardCompanyRequest {
	return companyentity.OnboardCompanyRequest{Name: "Acme Inc", Slug: "acme", AdminName: "Jane Admin", AdminEmail: "jane@acme.com", AdminPassword: "SuperSecurePassword123"}
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
	var d company_domainentity.OnboardingCompanyDomain
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
	var d company_domainentity.OnboardingCompanyDomain
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
	db.Model(&company_domainentity.OnboardingCompanyDomain{}).Where("domain = ?", "dreammakers.kilashop.com").Count(&count)
	if count != 0 {
		t.Fatalf("expected technical domain skipped, got %d", count)
	}
}

func TestService_Onboard_DuplicateDomain_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	_ = db.Create(&companyentity.OnboardingCompany{Slug: "existing", Name: "Existing", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&company_domainentity.OnboardingCompanyDomain{CompanyID: 1, Domain: "dreammakers.com.co", Kind: core.DomainKindPrimary, Status: core.DomainStatusActive}).Error
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
	_ = db.Create(&companyentity.OnboardingCompany{Slug: "existing", Name: "Existing", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&company_domainentity.OnboardingCompanyDomain{CompanyID: 1, Domain: "dreammakers.com.co", Kind: core.DomainKindPrimary, Status: core.DomainStatusActive}).Error
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
	_ = db.Create(&companyentity.OnboardingCompany{Slug: "existing", Name: "Existing", Status: core.CompanyStatusActive}).Error
	_ = db.Create(&company_domainentity.OnboardingCompanyDomain{CompanyID: 1, Domain: "dreammakers.kilashop.com", Kind: core.DomainKindTechnical, Status: core.DomainStatusActive}).Error
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
	_ = db.Create(&companyentity.OnboardingCompany{Name: "Acme Existing", Slug: "acme", Status: core.CompanyStatusActive}).Error
	_, err := svc.Onboard(context.Background(), baseRequest())
	if !errors.Is(err, core.ErrDuplicateCompanySlug) {
		t.Fatalf("expected duplicate slug, got: %v", err)
	}
}

func TestService_Onboard_DuplicateEmail_Rollback(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	svc := newService(db, "")
	_ = db.Create(&userentity.OnboardingUser{Name: "Jane Existing", Email: "jane@acme.com", PasswordHash: "secret", RoleID: 1}).Error
	req := baseRequest()
	req.Slug = "acme-new"
	_, err := svc.Onboard(context.Background(), req)
	if !errors.Is(err, core.ErrDuplicateAdminEmail) {
		t.Fatalf("expected duplicate email, got: %v", err)
	}
}

// countingHasher verifies conflicts are rejected before credential provisioning.
type countingHasher struct {
	calls int
	err   error
}

func (h *countingHasher) HashPassword(string) (string, error) {
	h.calls++
	return "hash", h.err
}

func onboardingCounts(t *testing.T, db *gorm.DB) []int64 {
	t.Helper()
	counts := make([]int64, 6)
	for i, table := range []string{"companies", "company_domains", "roles", "users", "permissions", "role_permissions"} {
		if err := db.Table(table).Count(&counts[i]).Error; err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
	}
	return counts
}

func assertOnboardingCounts(t *testing.T, db *gorm.DB, want []int64) {
	t.Helper()
	got := onboardingCounts(t, db)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("table index %d: count %d, want %d", i, got[i], want[i])
		}
	}
}

func TestService_Onboard_SoftDeletedKeysPreventProvisioning(t *testing.T) {
	for _, tc := range []struct {
		name string
		want error
	}{
		{"slug", core.ErrDuplicateCompanySlug},
		{"primary domain", core.ErrDuplicateCompanyDomain},
		{"technical domain", core.ErrDuplicateTechnicalDomain},
		{"email", core.ErrDuplicateAdminEmail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			seedPermissions(t, db)
			req := baseRequest()
			deleted := shared.BaseModel{PublicID: "reserved_record", DeletedAt: gorm.DeletedAt{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}}
			var record any
			switch tc.name {
			case "slug":
				record = &companyentity.OnboardingCompany{BaseModel: deleted, Name: "Reserved", Slug: req.Slug, Status: core.CompanyStatusActive}
			case "primary domain", "technical domain":
				domain := "acme.example.com"
				kind := core.DomainKindPrimary
				if tc.name == "primary domain" {
					req.Domain = &domain
				} else {
					req.GenerateTechnicalDomain = true
					kind = core.DomainKindTechnical
				}
				record = &company_domainentity.OnboardingCompanyDomain{BaseModel: deleted, CompanyID: 99, Domain: domain, Kind: kind, Status: core.DomainStatusActive}
			case "email":
				record = &userentity.OnboardingUser{BaseModel: deleted, Name: "Reserved", Email: req.AdminEmail, PasswordHash: "hash", RoleID: 99}
			}
			if err := db.Create(record).Error; err != nil {
				t.Fatalf("seed reserved key: %v", err)
			}
			before := onboardingCounts(t, db)
			hasher := &countingHasher{}
			res, err := NewService(NewRepository(db), hasher, "example.com").Onboard(context.Background(), req)
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
			if res != nil {
				t.Errorf("conflict returned successful response: %#v", res)
			}
			if hasher.calls != 0 {
				t.Errorf("conflict reached password hashing %d times", hasher.calls)
			}
			assertOnboardingCounts(t, db, before)
		})
	}
}

func TestService_Onboard_HashFailureRollsBackProvisioning(t *testing.T) {
	db := setupTestDB(t)
	seedPermissions(t, db)
	before := onboardingCounts(t, db)
	failure := errors.New("test hashing failure")
	hasher := &countingHasher{err: failure}
	req := baseRequest()
	domain := "acme.org"
	req.Domain = &domain
	req.GenerateTechnicalDomain = true
	res, err := NewService(NewRepository(db), hasher, "example.com").Onboard(context.Background(), req)
	if !errors.Is(err, failure) || res != nil {
		t.Fatalf("expected propagated hash failure and nil response, got %v, %v", res, err)
	}
	if hasher.calls != 1 {
		t.Fatalf("expected late hashing failure, calls=%d", hasher.calls)
	}
	assertOnboardingCounts(t, db, before)
}
