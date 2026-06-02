package view_user

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Service orchestrates the view_user use case.
type Service interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*core.UserResponse, error)
}

type service struct{ repo Repository }

// NewService creates a view_user service backed by the given repository.
func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) GetByPublicID(tc tenant.TenantContext, publicID string) (*core.UserResponse, error) {
	return s.repo.GetByPublicID(tc, publicID)
}
