package update_permission

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
)

type Service interface {
	Update(publicID string, req permission.UpdatePermissionRequest) (*permission.PermissionResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Update(publicID string, req permission.UpdatePermissionRequest) (*permission.PermissionResponse, error) {
	model, err := s.repo.GetPermissionByPublicID(publicID)
	if err != nil {
		return nil, err
	}
	if model.IsSystem {
		return nil, core.ErrSystemImmutable
	}

	model.Name = req.Name
	model.Description = req.Description
	model.DisplayOrder = req.DisplayOrder

	if err := s.repo.UpdatePermission(model); err != nil {
		return nil, err
	}

	var response permission.PermissionResponse
	response.FromModel(*model)
	return &response, nil
}
