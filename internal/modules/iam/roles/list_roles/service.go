package list_roles

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]core.RoleResponse, int64, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) List(tc tenant.TenantContext, params query.ListParams) ([]core.RoleResponse, int64, error) {
	items, err := s.repo.List(tc, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(tc)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
