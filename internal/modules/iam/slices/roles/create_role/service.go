package create_role

import (
	"sort"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	Create(tc tenant.TenantContext, req iamrole.CreateRoleRequest) (*iamrole.RoleResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Create(tc tenant.TenantContext, req iamrole.CreateRoleRequest) (*iamrole.RoleResponse, error) {
	if roles.IsReservedRoleIdentity(req.Name, req.Slug) {
		return nil, core.ErrReservedRoleIdentity
	}
	if ok, err := s.repo.ExistsRoleByName(tc, req.Name); err != nil {
		return nil, err
	} else if ok {
		return nil, core.ErrRoleNameAlreadyExists
	}
	if ok, err := s.repo.ExistsRoleBySlug(tc, req.Slug); err != nil {
		return nil, err
	} else if ok {
		return nil, core.ErrRoleSlugAlreadyExists
	}
	item, err := s.repo.Create(tc, req)
	if err != nil {
		return nil, err
	}
	return toRoleResponse(item), nil
}

func toRoleResponse(r *iamrole.IAMRole) *iamrole.RoleResponse {
	perms := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, p.Slug)
	}
	sort.Strings(perms)
	var response iamrole.RoleResponse
	response.FromModel(r, perms)
	return &response
}
