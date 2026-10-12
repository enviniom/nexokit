package view_permission

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
)

type Service interface {
	GetByPublicID(publicID string) (*permission.PermissionResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) GetByPublicID(publicID string) (*permission.PermissionResponse, error) {
	model, err := s.repo.GetPermissionByPublicID(publicID)
	if err != nil {
		return nil, err
	}

	var response permission.PermissionResponse
	response.FromModel(*model)
	return &response, nil
}
