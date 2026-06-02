package assign_role_to_user

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Service orchestrates the assign_role_to_user use case.
type Service interface {
	ChangeRole(tc tenant.TenantContext, targetPublicID, actorPublicID string, req core.ChangeUserRoleRequest) (*core.UserResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates an assign_role_to_user service backed by the given repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ChangeRole(tc tenant.TenantContext, targetPublicID, actorPublicID string, req core.ChangeUserRoleRequest) (*core.UserResponse, error) {
	if actorPublicID == "" || actorPublicID == targetPublicID {
		return nil, core.ErrForbidden
	}

	targetUser, err := s.repo.GetUserByPublicID(tc, targetPublicID)
	if err != nil {
		return nil, err
	}

	rootRole, err := s.repo.GetRoleBySlug(core.RootRoleSlug)
	if err != nil {
		return nil, err
	}
	if targetUser.RoleID == rootRole.ID {
		return nil, core.ErrForbidden
	}

	targetRole, err := s.repo.GetAssignableRoleByPublicID(tc, req.RoleID)
	if err != nil {
		return nil, err
	}
	if targetRole.Slug == core.RootRoleSlug {
		return nil, core.ErrForbiddenRoleAssignment
	}
	if targetUser.CompanyID != nil && targetRole.CompanyID != nil && *targetUser.CompanyID != *targetRole.CompanyID {
		return nil, core.ErrInvalidCompanyScope
	}

	return s.repo.AssignRole(tc, targetUser, targetRole.ID)
}
