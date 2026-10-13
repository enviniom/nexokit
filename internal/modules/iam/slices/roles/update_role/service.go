package update_role

import (
	"sort"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	Update(tc tenant.TenantContext, publicID string, req iamrole.UpdateRoleRequest) (*iamrole.RoleResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Update(tc tenant.TenantContext, publicID string, req iamrole.UpdateRoleRequest) (*iamrole.RoleResponse, error) {
	role, err := s.repo.GetRoleByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, core.ErrRoleProtected
	}
	if roles.IsReservedRoleIdentity(role.Name, role.Slug) {
		return nil, core.ErrRoleProtected
	}
	if roles.IsReservedRoleIdentity(req.Name, req.Slug) {
		return nil, core.ErrReservedRoleIdentity
	}
	if ok, err := s.repo.ExistsRoleByName(tc, req.Name, role.ID); err != nil {
		return nil, err
	} else if ok {
		return nil, core.ErrRoleNameAlreadyExists
	}
	if ok, err := s.repo.ExistsRoleBySlug(tc, req.Slug, role.ID); err != nil {
		return nil, err
	} else if ok {
		return nil, core.ErrRoleSlugAlreadyExists
	}
	role.Name, role.Slug, role.Description = req.Name, req.Slug, req.Description
	if err := s.repo.UpdateRole(tc, role); err != nil {
		return nil, err
	}
	return toRoleResponse(role), nil
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
