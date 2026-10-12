package delete_user

import (
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Service orchestrates the delete_user use case.
type Service interface {
	Delete(tc tenant.TenantContext, publicID string) error
}

type service struct{ repo Repository }

// NewService creates a delete_user service backed by the given repository.
func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Delete(tc tenant.TenantContext, publicID string) error {
	return s.repo.Delete(tc, publicID)
}
