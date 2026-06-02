package view_role

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	View(tc tenant.TenantContext, publicID string) (*core.RoleResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) View(tc tenant.TenantContext, publicID string) (*core.RoleResponse, error) {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	return role, nil
}
