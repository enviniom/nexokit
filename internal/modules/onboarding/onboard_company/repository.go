package onboard_company

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/modules/onboarding/queries"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type Repository interface {
	EnsureCompanySlugAvailable(tx *gorm.DB, slug string) error
	EnsureDomainAvailable(tx *gorm.DB, domain string, duplicateErr error) error
	EnsureEmailAvailable(tx *gorm.DB, email string) error
	CreateCompany(tx *gorm.DB, name, slug string) (*core.OnboardingCompany, error)
	CreateCompanyDomain(tx *gorm.DB, companyID uint, domain, kind string) error
	CreateRole(tx *gorm.DB, companyID uint, name, slug, description string) (*core.OnboardingRole, error)
	ListSystemPermissions(tx *gorm.DB) ([]core.OnboardingPermission, error)
	AssignPermissionToRole(tx *gorm.DB, roleID, permissionID uint) error
	CreateAdminUser(tx *gorm.DB, companyID, roleID uint, name, email, passwordHash string) (*core.OnboardingUser, error)
}

type repository struct{}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) EnsureCompanySlugAvailable(tx *gorm.DB, slug string) error {
	return queries.CheckSlugAvailable(tx, slug)
}

func (r *repository) EnsureDomainAvailable(tx *gorm.DB, domain string, duplicateErr error) error {
	return queries.CheckDomainAvailable(tx, domain, duplicateErr)
}

func (r *repository) EnsureEmailAvailable(tx *gorm.DB, email string) error {
	return queries.CheckEmailAvailable(tx, email)
}

func (r *repository) CreateCompany(tx *gorm.DB, name, slug string) (*core.OnboardingCompany, error) {
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

	if err := tx.Create(company).Error; err != nil {
		return nil, err
	}

	return company, nil
}

func (r *repository) CreateCompanyDomain(tx *gorm.DB, companyID uint, domain, kind string) error {
	publicID, err := identity.Generate()
	if err != nil {
		return err
	}

	return tx.Create(&core.OnboardingCompanyDomain{
		BaseModel:         shared.BaseModel{PublicID: publicID},
		CompanyID:         companyID,
		Domain:            domain,
		Kind:              kind,
		Status:            core.DomainStatusActive,
		RedirectToPrimary: false,
	}).Error
}

func (r *repository) CreateRole(tx *gorm.DB, companyID uint, name, slug, description string) (*core.OnboardingRole, error) {
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

	if err := tx.Create(role).Error; err != nil {
		return nil, err
	}

	return role, nil
}

func (r *repository) ListSystemPermissions(tx *gorm.DB) ([]core.OnboardingPermission, error) {
	return queries.ListSystemPermissions(tx)
}

func (r *repository) AssignPermissionToRole(tx *gorm.DB, roleID, permissionID uint) error {
	return queries.AssignPermissionToRole(tx, roleID, permissionID)
}

func (r *repository) CreateAdminUser(tx *gorm.DB, companyID, roleID uint, name, email, passwordHash string) (*core.OnboardingUser, error) {
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

	if err := tx.Create(adminUser).Error; err != nil {
		return nil, err
	}

	return adminUser, nil
}
