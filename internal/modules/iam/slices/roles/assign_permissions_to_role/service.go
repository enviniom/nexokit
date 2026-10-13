package assign_permissions_to_role

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"strings"
)

type Service interface {
	Assign(tc tenant.TenantContext, publicID string, req iamrole.AssignRolePermissionsRequest) (*iamrole.RolePermissionAssignmentResponse, error)
}
type service struct {
	repo  Repository
	cache cache.Cache
}

func NewService(repo Repository, c cache.Cache) Service {
	return &service{repo: repo, cache: c}
}
func (s *service) Assign(tc tenant.TenantContext, publicID string, req iamrole.AssignRolePermissionsRequest) (*iamrole.RolePermissionAssignmentResponse, error) {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	// Root and admin authority is implicit; other identities have editable grants.
	name, slug := strings.TrimSpace(role.Name), strings.TrimSpace(role.Slug)
	if strings.EqualFold(name, roles.RootRoleSlug) || strings.EqualFold(slug, roles.RootRoleSlug) ||
		strings.EqualFold(name, roles.AdminRoleSlug) || strings.EqualFold(slug, roles.AdminRoleSlug) {
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
	if err := s.repo.ReplacePermissions(tc, role.ID, ids); err != nil {
		return nil, err
	}
	if err := s.repo.InvalidateRoleMemberPermissionCache(tc, role.ID, s.cache); err != nil {
		return nil, err
	}
	var response iamrole.RolePermissionAssignmentResponse
	response.FromSelection(role.PublicID, normalized, s.repo.BuildRolePermissionCatalog(catalog, selected))
	return &response, nil
}
