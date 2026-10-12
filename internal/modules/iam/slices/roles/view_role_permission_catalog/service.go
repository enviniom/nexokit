package view_role_permission_catalog

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	View(tc tenant.TenantContext, publicID string) ([]iamrole.RolePermissionGroupResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) View(tc tenant.TenantContext, publicID string) ([]iamrole.RolePermissionGroupResponse, error) {
	role, err := s.repo.GetRoleByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	catalog, err := s.repo.ListPermissionCatalog()
	if err != nil {
		return nil, err
	}
	granted := map[string]bool{}
	for _, p := range role.Permissions {
		granted[p.Slug] = true
	}
	return buildRolePermissionCatalog(catalog, granted), nil
}

func buildRolePermissionCatalog(catalog []permission.IAMPermission, granted map[string]bool) []iamrole.RolePermissionGroupResponse {
	return iamrole.BuildPermissionCatalog(catalog, granted)
}
