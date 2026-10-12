package resolve_role_by_slug

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
)

type Service interface {
	ResolveRoleBySlug(slug string) (*iamrole.IAMRole, error)
}

type Repository interface {
	GetRoleBySlug(slug string) (*iamrole.IAMRole, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ResolveRoleBySlug(slug string) (*iamrole.IAMRole, error) {
	return s.repo.GetRoleBySlug(slug)
}
