package list_roles

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"sort"
)

type Service interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]iamrole.RoleResponse, int64, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) List(tc tenant.TenantContext, params query.ListParams) ([]iamrole.RoleResponse, int64, error) {
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

func toRoleResponse(r *iamrole.IAMRole) iamrole.RoleResponse {
	perms := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, p.Slug)
	}
	sort.Strings(perms)
	var response iamrole.RoleResponse
	response.FromModel(r, perms)
	return response
}
