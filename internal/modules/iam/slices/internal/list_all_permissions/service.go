package list_all_permissions

import "github.com/enviniom/nexokit/internal/modules/iam/core/permission"

type Service interface {
	ListAllPermissions() ([]permission.IAMPermission, error)
}

type Repository interface {
	ListAllPermissions() ([]permission.IAMPermission, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ListAllPermissions() ([]permission.IAMPermission, error) {
	return s.repo.ListAllPermissions()
}
