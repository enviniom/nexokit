package onboard_company

import (
	"context"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"gorm.io/gorm"
)

type Service interface {
	Onboard(ctx context.Context, req core.OnboardCompanyRequest) (*core.OnboardCompanyResponse, error)
}

type service struct {
	db             *gorm.DB
	repo           Repository
	hasher         core.PasswordHasher
	platformDomain string
}

func NewService(db *gorm.DB, repo Repository, hasher core.PasswordHasher, platformDomain string) Service {
	return &service{db: db, repo: repo, hasher: hasher, platformDomain: normalizeDomain(platformDomain)}
}

func (s *service) Onboard(ctx context.Context, req core.OnboardCompanyRequest) (*core.OnboardCompanyResponse, error) {
	if errs := req.Validate(); len(errs) > 0 {
		return nil, apperror.ErrValidation
	}

	var res core.OnboardCompanyResponse

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		normalizedSlug := strings.ToLower(strings.TrimSpace(req.Slug))
		if err := s.repo.EnsureCompanySlugAvailable(tx, normalizedSlug); err != nil {
			return err
		}

		var normalizedDomain string
		if req.Domain != nil && strings.TrimSpace(*req.Domain) != "" {
			normalizedDomain = normalizeDomain(*req.Domain)
			if err := s.repo.EnsureDomainAvailable(tx, normalizedDomain, core.ErrDuplicateCompanyDomain); err != nil {
				return err
			}
		}

		var technicalDomain string
		if req.GenerateTechnicalDomain {
			if s.platformDomain == "" {
				return core.ErrMissingPlatformDomain
			}
			technicalDomain = normalizedSlug + "." + s.platformDomain
			if normalizedDomain != "" && normalizedDomain == technicalDomain {
				return core.ErrDuplicateTechnicalDomain
			}
			if err := s.repo.EnsureDomainAvailable(tx, technicalDomain, core.ErrDuplicateTechnicalDomain); err != nil {
				return err
			}
		}

		normalizedEmail := strings.ToLower(strings.TrimSpace(req.AdminEmail))
		if err := s.repo.EnsureEmailAvailable(tx, normalizedEmail); err != nil {
			return err
		}

		company, err := s.repo.CreateCompany(tx, req.Name, normalizedSlug)
		if err != nil {
			return err
		}

		if normalizedDomain != "" {
			if err := s.repo.CreateCompanyDomain(tx, company.ID, normalizedDomain, core.DomainKindPrimary); err != nil {
				return err
			}
		}
		if technicalDomain != "" {
			if err := s.repo.CreateCompanyDomain(tx, company.ID, technicalDomain, core.DomainKindTechnical); err != nil {
				return err
			}
		}

		adminRole, err := s.repo.CreateRole(tx, company.ID, "Admin", core.RoleSlugAdmin, "Tenant administrator with full capabilities")
		if err != nil {
			return err
		}

		userRole, err := s.repo.CreateRole(tx, company.ID, "User", core.RoleSlugUser, "Standard tenant user")
		if err != nil {
			return err
		}

		systemPermissions, err := s.repo.ListSystemPermissions(tx)
		if err != nil {
			return err
		}

		for _, perm := range systemPermissions {
			if err := s.repo.AssignPermissionToRole(tx, adminRole.ID, perm.ID); err != nil {
				return err
			}
		}

		for _, perm := range systemPermissions {
			if perm.Slug == "users.view" || perm.Slug == "roles.view" {
				if err := s.repo.AssignPermissionToRole(tx, userRole.ID, perm.ID); err != nil {
					return err
				}
			}
		}

		passwordHash, err := s.hasher.HashPassword(req.AdminPassword)
		if err != nil {
			return err
		}

		adminUser, err := s.repo.CreateAdminUser(tx, company.ID, adminRole.ID, req.AdminName, normalizedEmail, passwordHash)
		if err != nil {
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

func normalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
}
