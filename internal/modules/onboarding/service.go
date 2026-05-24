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
	ErrDuplicateCompanySlug      = errors.New("company slug already exists")
	ErrDuplicateCompanyDomain    = errors.New("company domain already exists")
	ErrDuplicateCompanySubdomain = errors.New("company subdomain already exists")
	ErrDuplicateAdminEmail       = errors.New("admin email already exists")
)

// Service defines company onboarding business operations.
type Service interface {
	Onboard(ctx context.Context, req OnboardCompanyRequest) (*OnboardCompanyResponse, error)
}

type service struct {
	db     *gorm.DB
	hasher users.PasswordHasher
}

// NewService creates a new onboarding service instance.
func NewService(db *gorm.DB, hasher users.PasswordHasher) Service {
	return &service{db: db, hasher: hasher}
}

// Onboard runs the complete transactional onboarding flow for a tenant.
func (s *service) Onboard(ctx context.Context, req OnboardCompanyRequest) (*OnboardCompanyResponse, error) {
	if errs := req.Validate(); len(errs) > 0 {
		return nil, apperror.ErrValidation
	}

	var res OnboardCompanyResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Validate company slug system-wide uniqueness
		normalizedSlug := strings.ToLower(strings.TrimSpace(req.Slug))
		var count int64
		if err := tx.Model(&companies.Company{}).Where("slug = ?", normalizedSlug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateCompanySlug
		}

		// 2. Validate domain uniqueness system-wide if provided
		var normalizedDomain *string
		if req.Domain != nil && *req.Domain != "" {
			d := strings.ToLower(strings.TrimSpace(*req.Domain))
			normalizedDomain = &d
			if err := tx.Model(&companies.Company{}).Where("domain = ?", *normalizedDomain).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDuplicateCompanyDomain
			}
		}

		// 3. Validate subdomain uniqueness system-wide if provided
		var normalizedSubdomain *string
		if req.Subdomain != nil && *req.Subdomain != "" {
			sd := strings.ToLower(strings.TrimSpace(*req.Subdomain))
			normalizedSubdomain = &sd
			if err := tx.Model(&companies.Company{}).Where("subdomain = ?", *normalizedSubdomain).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDuplicateCompanySubdomain
			}
		}

		// 4. Validate admin email uniqueness system-wide
		normalizedEmail := strings.ToLower(strings.TrimSpace(req.AdminEmail))
		if err := tx.Model(&users.User{}).Where("email = ?", normalizedEmail).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateAdminEmail
		}

		// 5. Create Company
		companyPubID, err := identity.Generate()
		if err != nil {
			return err
		}
		company := &companies.Company{
			BaseModel: shared.BaseModel{PublicID: companyPubID},
			Name:      req.Name,
			Slug:      normalizedSlug,
			Domain:    normalizedDomain,
			Subdomain: normalizedSubdomain,
			Status:    companies.CompanyStatusActive,
		}
		if err := tx.Create(company).Error; err != nil {
			return err
		}

		// 6. Create Tenant "admin" Role
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

		// 7. Create Tenant "user" Role
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

		// 8. Load All Registered System Permissions
		var systemPermissions []permissions.Permission
		if err := tx.Find(&systemPermissions).Error; err != nil {
			return err
		}

		// 9. Assign All System Permissions to Tenant "admin" Role
		for _, perm := range systemPermissions {
			if err := tx.Table("role_permissions").Create(map[string]any{
				"role_id":       adminRole.ID,
				"permission_id": perm.ID,
			}).Error; err != nil {
				return err
			}
		}

		// 10. Assign Base Permissions subset (users.view, roles.view) to Tenant "user" Role
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

		// 11. Hash Administrator Password
		passwordHash, err := s.hasher.HashPassword(req.AdminPassword)
		if err != nil {
			return err
		}

		// 12. Create Initial Administrator User
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

		// Save response values
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
