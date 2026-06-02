package update_user

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Service orchestrates the update_user use case.
type Service interface {
	Update(tc tenant.TenantContext, publicID, actorPublicID string, req core.UpdateUserRequest) (*core.UserResponse, error)
}

type service struct{ repo Repository }

// NewService creates an update_user service backed by the given repository.
func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Update(tc tenant.TenantContext, publicID, actorPublicID string, req core.UpdateUserRequest) (*core.UserResponse, error) {
	u, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}

	rootRole, err := s.repo.GetRoleBySlug(core.RootRoleSlug)
	if err != nil {
		return nil, err
	}

	if u.RoleID == rootRole.ID {
		if actorPublicID == "" || actorPublicID != u.PublicID {
			return nil, core.ErrForbiddenRoleAssignment
		}
		u.Name = req.Name
		u.Email = req.Email
	} else {
		if !tc.IsRootScope {
			if req.CompanyID != nil && *req.CompanyID != tc.CompanyID {
				return nil, core.ErrForbiddenRoleAssignment
			}
			cid := tc.CompanyID
			req.CompanyID = &cid
		} else if req.CompanyID == nil {
			return nil, core.ErrInvalidCompanyScope
		}
		u.Name = req.Name
		u.Email = req.Email
		u.CompanyID = req.CompanyID
	}

	if err := s.repo.Update(u); err != nil {
		return nil, err
	}

	return s.repo.Reload(tc, u.PublicID)
}
