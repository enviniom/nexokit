package resolve_role_by_slug

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
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
	item, err := s.repo.GetRoleBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}
