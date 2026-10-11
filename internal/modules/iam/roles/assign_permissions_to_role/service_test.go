package assign_permissions_to_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepository struct {
	role              *core.IAMRole
	roleErr           error
	catalog           []core.IAMPermission
	catalogErr        error
	selected          map[string]bool
	normalized        []string
	ids               []uint
	selectionErr      error
	removeSystem      bool
	replaceErr        error
	invalidateErr     error
	replacedTenant    tenant.TenantContext
	invalidatedTenant tenant.TenantContext
	replacedRoleID    uint
	replacedPermIDs   []uint
}

func (f *fakeRepository) GetByPublicID(tenant.TenantContext, string) (*core.IAMRole, error) {
	if f.roleErr != nil {
		return nil, f.roleErr
	}
	return f.role, nil
}
func (f *fakeRepository) ListAllPermissions() ([]core.IAMPermission, error) {
	if f.catalogErr != nil {
		return nil, f.catalogErr
	}
	return f.catalog, nil
}
func (f *fakeRepository) ReplacePermissions(tc tenant.TenantContext, roleID uint, permissionIDs []uint) error {
	f.replacedTenant = tc
	f.replacedRoleID = roleID
	f.replacedPermIDs = permissionIDs
	return f.replaceErr
}
func (f *fakeRepository) InvalidateRoleMemberPermissionCache(tc tenant.TenantContext, roleID uint, _ cache.Cache) error {
	f.invalidatedTenant = tc
	return f.invalidateErr
}
func (f *fakeRepository) ResolvePermissionSelection([]core.IAMPermission, []string) ([]string, map[string]bool, []uint, error) {
	if f.selectionErr != nil {
		return nil, nil, nil, f.selectionErr
	}
	return f.normalized, f.selected, f.ids, nil
}
func (f *fakeRepository) RemovesSystemPermission([]core.IAMPermission, map[string]bool) bool {
	return f.removeSystem
}
func (f *fakeRepository) BuildRolePermissionCatalog([]core.IAMPermission, map[string]bool) []core.RolePermissionGroupResponse {
	return []core.RolePermissionGroupResponse{{Module: "roles"}}
}

func TestServiceAssign(t *testing.T) {
	errRepo := errors.New("repo")
	baseRole := &core.IAMRole{BaseModel: shared.BaseModel{ID: 11, PublicID: "role-11"}, Slug: "manager"}

	tests := []struct {
		name    string
		repo    *fakeRepository
		actor   []string
		wantErr error
	}{
		{name: "forbidden without permission", repo: &fakeRepository{role: baseRole}, actor: []string{"users.read"}, wantErr: core.ErrRoleProtected},
		{name: "role not found", repo: &fakeRepository{roleErr: core.ErrNotFound}, actor: []string{"roles.assign_permissions"}, wantErr: core.ErrNotFound},
		{name: "immutable admin role", repo: &fakeRepository{role: &core.IAMRole{Slug: core.AdminRoleSlug}}, actor: []string{"roles.assign_permissions"}, wantErr: core.ErrSystemImmutable},
		{name: "selection error propagated", repo: &fakeRepository{role: baseRole, selectionErr: core.ErrInvalidPermissionSlug}, actor: []string{"roles.assign_permissions"}, wantErr: core.ErrInvalidPermissionSlug},
		{name: "cannot remove system permission", repo: &fakeRepository{role: &core.IAMRole{IsSystem: true}, selected: map[string]bool{"roles.read": true}, ids: []uint{1}, removeSystem: true}, actor: []string{"roles.assign_permissions"}, wantErr: core.ErrSystemImmutable},
		{name: "replace error propagated", repo: &fakeRepository{role: baseRole, selected: map[string]bool{"roles.read": true}, ids: []uint{1}, replaceErr: errRepo}, actor: []string{"roles.assign_permissions"}, wantErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil)
			_, err := svc.Assign(tenant.NewRoot(), "role-11", core.AssignRolePermissionsRequest{Permissions: []string{"roles.read"}}, tt.actor)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	t.Run("success returns response and writes ids", func(t *testing.T) {
		repo := &fakeRepository{
			role:       baseRole,
			catalog:    []core.IAMPermission{{Slug: "roles.read"}},
			normalized: []string{"roles.read"},
			selected:   map[string]bool{"roles.read": true},
			ids:        []uint{7},
		}
		svc := NewService(repo, nil)

		tc := tenant.NewScoped(7, "selected-company")
		res, err := svc.Assign(tc, "role-11", core.AssignRolePermissionsRequest{Permissions: []string{"roles.read"}}, []string{"roles.assign_permissions"})
		if repo.replacedTenant != tc || repo.invalidatedTenant != tc {
			t.Fatal("trusted tenant context was not forwarded to persistence and cache")
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.RoleID != "role-11" {
			t.Fatalf("expected role-11, got %s", res.RoleID)
		}
		if repo.replacedRoleID != 11 {
			t.Fatalf("expected replaced role id 11, got %d", repo.replacedRoleID)
		}
	})
}
