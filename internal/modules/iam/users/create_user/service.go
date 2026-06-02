package create_user

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

// PasswordHasher abstracts password hashing for the create_user slice.
type PasswordHasher interface {
	HashPassword(string) (string, error)
}

// Service orchestrates the create_user use case.
type Service interface {
	Create(tc tenant.TenantContext, req core.CreateUserRequest) (*core.UserResponse, error)
}

type service struct {
	repo   Repository
	hasher PasswordHasher
}

// NewService creates a create_user service backed by the given repository and hasher.
func NewService(repo Repository, hasher PasswordHasher) Service {
	return &service{repo: repo, hasher: hasher}
}

func (s *service) Create(tc tenant.TenantContext, req core.CreateUserRequest) (*core.UserResponse, error) {
	rootRole, err := s.repo.GetRoleBySlug(core.RootRoleSlug)
	if err != nil {
		return nil, err
	}

	if req.RoleID == rootRole.ID {
		if !tc.IsRootScope || req.CompanyID != nil {
			return nil, core.ErrForbiddenRoleAssignment
		}
	} else if tc.IsRootScope {
		if req.CompanyID == nil {
			return nil, core.ErrInvalidCompanyScope
		}
	} else {
		if req.CompanyID != nil && *req.CompanyID != tc.CompanyID {
			return nil, core.ErrForbiddenRoleAssignment
		}
		cid := tc.CompanyID
		req.CompanyID = &cid
	}

	if exists, err := s.repo.ExistsByEmail(req.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, core.ErrUserEmailAlreadyExists
	}

	pid, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	hash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: pid},
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		RoleID:       req.RoleID,
		CompanyID:    req.CompanyID,
		IsActive:     true,
	}

	if err := s.repo.Create(u); err != nil {
		return nil, err
	}

	return s.repo.GetByPublicID(tenant.NewRoot(), u.PublicID)
}
