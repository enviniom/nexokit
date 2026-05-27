package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

var (
	ErrDuplicateCompanySlug     = errors.New("company slug already exists")
	ErrDuplicateCompanyDomain   = errors.New("company domain already exists")
	ErrDuplicateTechnicalDomain = errors.New("company technical domain already exists")
	ErrMissingPlatformDomain    = errors.New("platform domain is required to generate technical domain")
	ErrDuplicateAdminEmail      = errors.New("admin email already exists")
)

// Service defines company onboarding business operations.
type Service interface {
	Onboard(ctx context.Context, req OnboardCompanyRequest) (*OnboardCompanyResponse, error)
}

type service struct {
	db             *gorm.DB
	hasher         users.PasswordHasher
	platformDomain string
}

// Option configures onboarding service behavior.
type Option func(*service)

// WithPlatformDomain configures the base domain used for generated technical company domains.
func WithPlatformDomain(domain string) Option {
	return func(s *service) {
		s.platformDomain = normalizeDomain(domain)
	}
}

// NewService creates a new onboarding service instance.
func NewService(db *gorm.DB, hasher users.PasswordHasher, opts ...Option) Service {
	svc := &service{db: db, hasher: hasher}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// Onboard runs the complete transactional onboarding flow for a tenant.
func (s *service) Onboard(ctx context.Context, req OnboardCompanyRequest) (*OnboardCompanyResponse, error) {
	if errs := req.Validate(); len(errs) > 0 {
		return nil, apperror.ErrValidation
	}

	var res OnboardCompanyResponse

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		normalizedSlug := strings.ToLower(strings.TrimSpace(req.Slug))
		var count int64
		if err := tx.Model(&companies.Company{}).Where("slug = ?", normalizedSlug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateCompanySlug
		}

		var normalizedDomain string
		if req.Domain != nil && strings.TrimSpace(*req.Domain) != "" {
			normalizedDomain = normalizeDomain(*req.Domain)
			if err := ensureDomainAvailable(tx, normalizedDomain, ErrDuplicateCompanyDomain); err != nil {
				return err
			}
		}

		var technicalDomain string
		if req.GenerateTechnicalDomain {
			if s.platformDomain == "" {
				return ErrMissingPlatformDomain
			}
			technicalDomain = normalizedSlug + "." + s.platformDomain
			if normalizedDomain != "" && normalizedDomain == technicalDomain {
				return ErrDuplicateTechnicalDomain
			}
			if err := ensureDomainAvailable(tx, technicalDomain, ErrDuplicateTechnicalDomain); err != nil {
				return err
			}
		}

		normalizedEmail := strings.ToLower(strings.TrimSpace(req.AdminEmail))
		if err := tx.Model(&users.User{}).Where("email = ?", normalizedEmail).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateAdminEmail
		}

		companyPubID, err := identity.Generate()
		if err != nil {
			return err
		}
		company := &companies.Company{
			BaseModel: shared.BaseModel{PublicID: companyPubID},
			Name:      req.Name,
			Slug:      normalizedSlug,
			Status:    companies.CompanyStatusActive,
		}
		if err := tx.Create(company).Error; err != nil {
			return err
		}

		if normalizedDomain != "" {
			if err := createCompanyDomain(tx, company.ID, normalizedDomain, companies.CompanyDomainKindPrimary); err != nil {
				return err
			}
		}
		if technicalDomain != "" {
			if err := createCompanyDomain(tx, company.ID, technicalDomain, companies.CompanyDomainKindTechnical); err != nil {
				return err
			}
		}

		adminRolePubID, err := identity.Generate()
		if err != nil {
			return err
		}
		adminRole := &roles.Role{
			BaseModel:   shared.BaseModel{PublicID: adminRolePubID},
			Name:        "Admin",
			Slug:        roles.AdminRoleSlug,
			CompanyID:   &company.ID,
			Description: "Tenant administrator with full capabilities",
			IsSystem:    true,
		}
		if err := tx.Create(adminRole).Error; err != nil {
			return err
		}

		userRolePubID, err := identity.Generate()
		if err != nil {
			return err
		}
		userRole := &roles.Role{
			BaseModel:   shared.BaseModel{PublicID: userRolePubID},
			Name:        "User",
			Slug:        roles.UserRoleSlug,
			CompanyID:   &company.ID,
			Description: "Standard tenant user",
			IsSystem:    true,
		}
		if err := tx.Create(userRole).Error; err != nil {
			return err
		}

		var systemPermissions []permissions.Permission
		if err := tx.Find(&systemPermissions).Error; err != nil {
			return err
		}

		for _, perm := range systemPermissions {
			if err := tx.Table("role_permissions").Create(map[string]any{
				"role_id":       adminRole.ID,
				"permission_id": perm.ID,
			}).Error; err != nil {
				return err
			}
		}

		for _, perm := range systemPermissions {
			if perm.Slug == "users.view" || perm.Slug == "roles.view" {
				if err := tx.Table("role_permissions").Create(map[string]any{
					"role_id":       userRole.ID,
					"permission_id": perm.ID,
				}).Error; err != nil {
					return err
				}
			}
		}

		passwordHash, err := s.hasher.HashPassword(req.AdminPassword)
		if err != nil {
			return err
		}

		adminUserPubID, err := identity.Generate()
		if err != nil {
			return err
		}
		adminUser := &users.User{
			BaseModel:    shared.BaseModel{PublicID: adminUserPubID},
			Name:         req.AdminName,
			Email:        normalizedEmail,
			PasswordHash: passwordHash,
			RoleID:       adminRole.ID,
			CompanyID:    &company.ID,
			IsActive:     true,
		}
		if err := tx.Create(adminUser).Error; err != nil {
			return err
		}

		res.CompanyPublicID = company.PublicID
		res.CompanySlug = company.Slug
		res.AdminPublicID = adminUser.PublicID
		res.AdminEmail = adminUser.Email

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func ensureDomainAvailable(tx *gorm.DB, domain string, duplicateErr error) error {
	var count int64
	if err := tx.Model(&companies.CompanyDomain{}).Where("domain = ?", domain).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return duplicateErr
	}
	return nil
}

func createCompanyDomain(tx *gorm.DB, companyID uint, domain, kind string) error {
	publicID, err := identity.Generate()
	if err != nil {
		return err
	}
	return tx.Create(&companies.CompanyDomain{
		BaseModel:         shared.BaseModel{PublicID: publicID},
		CompanyID:         companyID,
		Domain:            domain,
		Kind:              kind,
		Status:            companies.CompanyDomainStatusActive,
		RedirectToPrimary: false,
	}).Error
}

func normalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
}
