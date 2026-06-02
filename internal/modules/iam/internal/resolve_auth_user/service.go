package resolve_auth_user

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"gorm.io/gorm"
)

type Service interface {
	ResolveAuthUser(publicID string) (*authctx.User, error)
}

type Repository interface {
	GetAuthUser(publicID string) (*core.IAMUser, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ResolveAuthUser(publicID string) (*authctx.User, error) {
	user, err := s.repo.GetAuthUser(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	slugs := make([]string, len(user.Role.Permissions))
	for i, p := range user.Role.Permissions {
		slugs[i] = p.Slug
	}
	return &authctx.User{
		ID:          user.ID,
		PublicID:    user.PublicID,
		Email:       user.Email,
		Name:        user.Name,
		Role:        user.Role.Name,
		RoleSlug:    user.Role.Slug,
		RoleID:      user.RoleID,
		CompanyID:   user.CompanyID,
		IsRoot:      user.Role.Slug == core.RootRoleSlug,
		IsActive:    user.IsActive,
		Permissions: slugs,
	}, nil
}
