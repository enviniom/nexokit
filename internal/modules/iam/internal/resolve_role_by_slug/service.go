package resolve_role_by_slug

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type Service interface {
	ResolveRoleBySlug(slug string) (*core.IAMRole, error)
}

type Repository interface {
	GetRoleBySlug(slug string) (*core.IAMRole, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ResolveRoleBySlug(slug string) (*core.IAMRole, error) {
	return s.repo.GetRoleBySlug(slug)
}
