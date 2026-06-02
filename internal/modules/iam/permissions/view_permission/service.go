package view_permission

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type Service interface {
	GetByPublicID(publicID string) (*core.PermissionResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) GetByPublicID(publicID string) (*core.PermissionResponse, error) {
	permission, err := s.repo.GetPermissionByPublicID(publicID)
	if err != nil {
		return nil, err
	}

	return &core.PermissionResponse{
		PublicID:     permission.PublicID,
		Slug:         permission.Slug,
		Name:         permission.Name,
		Module:       permission.Module,
		Action:       permission.Action,
		Description:  permission.Description,
		IsSystem:     permission.IsSystem,
		DisplayOrder: permission.DisplayOrder,
		CreatedAt:    permission.CreatedAt,
		UpdatedAt:    permission.UpdatedAt,
		CreatedBy:    permission.CreatedBy,
		UpdatedBy:    permission.UpdatedBy,
	}, nil
}
