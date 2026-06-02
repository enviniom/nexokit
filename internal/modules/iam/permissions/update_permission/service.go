package update_permission

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type Service interface {
	Update(publicID string, req core.UpdatePermissionRequest) (*core.PermissionResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Update(publicID string, req core.UpdatePermissionRequest) (*core.PermissionResponse, error) {
	permission, err := s.repo.GetPermissionByPublicID(publicID)
	if err != nil {
		return nil, err
	}
	if permission.IsSystem {
		return nil, core.ErrSystemImmutable
	}

	permission.Name = req.Name
	permission.Description = req.Description
	permission.DisplayOrder = req.DisplayOrder

	if err := s.repo.UpdatePermission(permission); err != nil {
		return nil, err
	}

	return toPermissionResponse(permission), nil
}
