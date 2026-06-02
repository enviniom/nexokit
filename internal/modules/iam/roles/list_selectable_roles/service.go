package list_selectable_roles

import (
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	List(tc tenant.TenantContext) ([]response.SelectResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) List(tc tenant.TenantContext) ([]response.SelectResponse, error) {
	items, err := s.repo.List(tc)
	if err != nil {
		return nil, err
	}
	return items, nil
}
