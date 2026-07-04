package onboard_company

import (
	"context"
	"fmt"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/platform/shared/string"
)

// Service coordinates the company onboarding workflow.
type Service interface {
	Onboard(ctx context.Context, req core.OnboardCompanyRequest) (*core.OnboardCompanyResponse, error)
}

type service struct {
	repo           Repository
	hasher         core.PasswordHasher
	platformDomain string
}

// NewService creates a new onboard_company service.
func NewService(repo Repository, hasher core.PasswordHasher, platformDomain string) Service {
	return &service{
		repo:           repo,
		hasher:         hasher,
		platformDomain: str.NormalizeDomain(platformDomain),
	}
}

func (s *service) Onboard(ctx context.Context, req core.OnboardCompanyRequest) (*core.OnboardCompanyResponse, error) {
	if errs := req.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errs)
	}

	var res core.OnboardCompanyResponse

	err := s.repo.WithTx(ctx, func(tx Repository) error {
		normalizedSlug := str.NormalizeSlug(req.Slug)
		if err := tx.EnsureCompanySlugAvailable(normalizedSlug); err != nil {
			return err
		}

		var normalizedDomain string
		if req.Domain != nil && str.NormalizeDomain(*req.Domain) != "" {
			normalizedDomain = str.NormalizeDomain(*req.Domain)
			if err := tx.EnsureDomainAvailable(normalizedDomain, core.ErrDuplicateCompanyDomain); err != nil {
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
			if err := tx.EnsureDomainAvailable(technicalDomain, core.ErrDuplicateTechnicalDomain); err != nil {
				return err
			}
		}

		normalizedEmail := str.NormalizeEmail(req.AdminEmail)
		if err := tx.EnsureEmailAvailable(normalizedEmail); err != nil {
			return err
		}

		company, err := tx.CreateCompany(req.Name, normalizedSlug)
		if err != nil {
			return err
		}

		if normalizedDomain != "" {
			if err := tx.CreateCompanyDomain(company.ID, normalizedDomain, core.DomainKindPrimary); err != nil {
				return err
			}
		}
		if technicalDomain != "" {
			if err := tx.CreateCompanyDomain(company.ID, technicalDomain, core.DomainKindTechnical); err != nil {
				return err
			}
		}

		adminRole, err := tx.CreateRole(company.ID, "Admin", core.RoleSlugAdmin, "Tenant administrator with full capabilities")
		if err != nil {
			return err
		}

		userRole, err := tx.CreateRole(company.ID, "User", core.RoleSlugUser, "Standard tenant user")
		if err != nil {
			return err
		}

		systemPermissions, err := tx.ListSystemPermissions()
		if err != nil {
			return err
		}

		for _, perm := range systemPermissions {
			if err := tx.AssignPermissionToRole(adminRole.ID, perm.ID); err != nil {
				return err
			}
		}

		for _, perm := range systemPermissions {
			if perm.Slug == "users.view" || perm.Slug == "roles.view" {
				if err := tx.AssignPermissionToRole(userRole.ID, perm.ID); err != nil {
					return err
				}
			}
		}

		passwordHash, err := s.hasher.HashPassword(req.AdminPassword)
		if err != nil {
			return err
		}

		adminUser, err := tx.CreateAdminUser(company.ID, adminRole.ID, req.AdminName, normalizedEmail, passwordHash)
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
