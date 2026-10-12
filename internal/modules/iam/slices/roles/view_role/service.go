package view_role

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"sort"
)

type Service interface {
	View(tc tenant.TenantContext, publicID string) (*iamrole.RoleResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) View(tc tenant.TenantContext, publicID string) (*iamrole.RoleResponse, error) {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func toRoleResponse(r *iamrole.IAMRole) *iamrole.RoleResponse {
	perms := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, p.Slug)
	}
	sort.Strings(perms)
	var response iamrole.RoleResponse
	response.FromViewModel(r, perms)
	return &response
}
