package list_all_permissions

import "github.com/enviniom/nexokit/internal/modules/iam/core"

type Service interface {
	ListAllPermissions() ([]core.IAMPermission, error)
}

type Repository interface {
	ListAllPermissions() ([]core.IAMPermission, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ListAllPermissions() ([]core.IAMPermission, error) {
	return s.repo.ListAllPermissions()
}
