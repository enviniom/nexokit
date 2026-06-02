package view_role_permission_catalog

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	View(tc tenant.TenantContext, publicID string) ([]core.RolePermissionGroupResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) View(tc tenant.TenantContext, publicID string) ([]core.RolePermissionGroupResponse, error) {
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

func buildRolePermissionCatalog(catalog []core.IAMPermission, granted map[string]bool) []core.RolePermissionGroupResponse {
	groups := []core.RolePermissionGroupResponse{}
	idx := map[string]int{}

	for _, p := range catalog {
		i, ok := idx[p.Module]
		if !ok {
			groups = append(groups, core.RolePermissionGroupResponse{Module: p.Module})
			i = len(groups) - 1
			idx[p.Module] = i
		}

		groups[i].Permissions = append(groups[i].Permissions, core.RolePermissionResponse{
			PublicID:     p.PublicID,
			Slug:         p.Slug,
			Name:         p.Name,
			Module:       p.Module,
			Action:       p.Action,
			Description:  p.Description,
			IsSystem:     p.IsSystem,
			DisplayOrder: p.DisplayOrder,
			Granted:      granted[p.Slug],
		})
	}

	return groups
}
