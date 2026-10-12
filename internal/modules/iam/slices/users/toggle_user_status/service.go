package toggle_user_status

import (
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Service orchestrates the toggle_user_status use case.
type Service interface {
	Toggle(tc tenant.TenantContext, publicID string, req iamuser.UpdateStatusRequest) (*iamuser.UserResponse, error)
}

type service struct{ repo Repository }

// NewService creates a toggle_user_status service backed by the given repository.
func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Toggle(tc tenant.TenantContext, publicID string, req iamuser.UpdateStatusRequest) (*iamuser.UserResponse, error) {
	return s.repo.ToggleStatus(tc, publicID, req.IsActive)
}
