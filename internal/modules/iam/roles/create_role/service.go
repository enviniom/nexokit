package create_role

import (
	"sort"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	Create(tc tenant.TenantContext, req core.CreateRoleRequest) (*core.RoleResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Create(tc tenant.TenantContext, req core.CreateRoleRequest) (*core.RoleResponse, error) {
	if core.IsReservedRoleIdentity(req.Name, req.Slug) {
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

func toRoleResponse(r *core.IAMRole) *core.RoleResponse {
	perms := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, p.Slug)
	}
	sort.Strings(perms)
	return &core.RoleResponse{PublicID: r.PublicID, CompanyID: companyPublicID(r), Name: r.Name, Slug: r.Slug, Description: r.Description, IsSystem: r.IsSystem, Permissions: perms, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, CreatedBy: r.CreatedBy, UpdatedBy: r.UpdatedBy}
}

func companyPublicID(r *core.IAMRole) *string {
	if r.Company.PublicID == "" {
		return nil
	}
	v := r.Company.PublicID
	return &v
}
