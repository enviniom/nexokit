package assign_permissions_to_role

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	Assign(tc tenant.TenantContext, publicID string, req core.AssignRolePermissionsRequest, actorPermissions []string) (*core.RolePermissionAssignmentResponse, error)
}
type service struct {
	repo  Repository
	cache cache.Cache
}

func NewService(repo Repository, c cache.Cache) Service {
	return &service{repo: repo, cache: c}
}
func (s *service) Assign(tc tenant.TenantContext, publicID string, req core.AssignRolePermissionsRequest, actorPermissions []string) (*core.RolePermissionAssignmentResponse, error) {
	if !hasPermission(actorPermissions, "roles.assign_permissions") {
		return nil, core.ErrRoleProtected
	}
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	if role.Slug == core.AdminRoleSlug {
		return nil, core.ErrSystemImmutable
	}
	catalog, err := s.repo.ListAllPermissions()
	if err != nil {
		return nil, err
	}
	normalized, selected, ids, err := s.repo.ResolvePermissionSelection(catalog, req.Permissions)
	if err != nil {
		return nil, err
	}
	if role.IsSystem && s.repo.RemovesSystemPermission(role.Permissions, selected) {
		return nil, core.ErrSystemImmutable
	}
	if err := s.repo.ReplacePermissions(role.ID, ids); err != nil {
		return nil, err
	}
	if err := s.repo.InvalidateRoleMemberPermissionCache(role.ID, s.cache); err != nil {
		return nil, err
	}
	return &core.RolePermissionAssignmentResponse{RoleID: role.PublicID, Permissions: normalized, Catalog: s.repo.BuildRolePermissionCatalog(catalog, selected)}, nil
}
func hasPermission(items []string, slug string) bool {
	for _, i := range items {
		if i == "*" || i == slug {
			return true
		}
	}
	return false
}
