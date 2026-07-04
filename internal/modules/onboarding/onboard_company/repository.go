package onboard_company

import (
	"context"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/modules/onboarding/queries"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// Repository defines persistence operations for the onboard_company slice.
type Repository interface {
	WithTx(ctx context.Context, fn func(tx Repository) error) error
	EnsureCompanySlugAvailable(slug string) error
	EnsureDomainAvailable(domain string, duplicateErr error) error
	EnsureEmailAvailable(email string) error
	CreateCompany(name, slug string) (*core.OnboardingCompany, error)
	CreateCompanyDomain(companyID uint, domain, kind string) error
	CreateRole(companyID uint, name, slug, description string) (*core.OnboardingRole, error)
	ListSystemPermissions() ([]core.OnboardingPermission, error)
	AssignPermissionToRole(roleID, permissionID uint) error
	CreateAdminUser(companyID, roleID uint, name, email, passwordHash string) (*core.OnboardingUser, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new onboarding repository backed by db.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// WithTx runs fn inside a GORM transaction. The repository passed to fn is
// bound to the transaction so every operation participates in the same unit.
func (r *repository) WithTx(ctx context.Context, fn func(tx Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repository{db: tx})
	})
}

func (r *repository) EnsureCompanySlugAvailable(slug string) error {
	return queries.CheckSlugAvailable(r.db, slug)
}

func (r *repository) EnsureDomainAvailable(domain string, duplicateErr error) error {
	return queries.CheckDomainAvailable(r.db, domain, duplicateErr)
}

func (r *repository) EnsureEmailAvailable(email string) error {
	return queries.CheckEmailAvailable(r.db, email)
}

func (r *repository) CreateCompany(name, slug string) (*core.OnboardingCompany, error) {
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	company := &core.OnboardingCompany{
		BaseModel: shared.BaseModel{PublicID: publicID},
		Name:      name,
		Slug:      slug,
		Status:    core.CompanyStatusActive,
	}

	if err := r.db.Create(company).Error; err != nil {
		if gormutil.IsUniqueConstraintError(err) {
			return nil, core.ErrDuplicateCompanySlug
		}
		return nil, err
	}

	return company, nil
}

func (r *repository) CreateCompanyDomain(companyID uint, domain, kind string) error {
	publicID, err := identity.Generate()
	if err != nil {
		return err
	}

	err = r.db.Create(&core.OnboardingCompanyDomain{
		BaseModel:         shared.BaseModel{PublicID: publicID},
		CompanyID:         companyID,
		Domain:            domain,
		Kind:              kind,
		Status:            core.DomainStatusActive,
		RedirectToPrimary: false,
	}).Error
	if err == nil {
		return nil
	}

	if gormutil.IsUniqueConstraintError(err) {
		if kind == core.DomainKindTechnical {
			return core.ErrDuplicateTechnicalDomain
		}
		return core.ErrDuplicateCompanyDomain
	}
	return err
}

func (r *repository) CreateRole(companyID uint, name, slug, description string) (*core.OnboardingRole, error) {
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	role := &core.OnboardingRole{
		BaseModel:   shared.BaseModel{PublicID: publicID},
		Name:        name,
		Slug:        slug,
		CompanyID:   &companyID,
		Description: description,
		IsSystem:    true,
	}

	if err := r.db.Create(role).Error; err != nil {
		return nil, err
	}

	return role, nil
}

func (r *repository) ListSystemPermissions() ([]core.OnboardingPermission, error) {
	return queries.ListSystemPermissions(r.db)
}

func (r *repository) AssignPermissionToRole(roleID, permissionID uint) error {
	return queries.AssignPermissionToRole(r.db, roleID, permissionID)
}

func (r *repository) CreateAdminUser(companyID, roleID uint, name, email, passwordHash string) (*core.OnboardingUser, error) {
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	adminUser := &core.OnboardingUser{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		RoleID:       roleID,
		CompanyID:    &companyID,
		IsActive:     true,
	}

	if err := r.db.Create(adminUser).Error; err != nil {
		if gormutil.IsUniqueConstraintError(err) {
			return nil, core.ErrDuplicateAdminEmail
		}
		return nil, err
	}

	return adminUser, nil
}
