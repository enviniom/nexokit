package assign_permissions_to_role

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"testing"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepository struct {
	role              *iamrole.IAMRole
	roleErr           error
	catalog           []permission.IAMPermission
	catalogErr        error
	selected          map[string]bool
	normalized        []string
	ids               []uint
	selectionErr      error
	replaceErr        error
	invalidateErr     error
	replacedTenant    tenant.TenantContext
	invalidatedTenant tenant.TenantContext
	replacedRoleID    uint
	replacedPermIDs   []uint
}

func (f *fakeRepository) GetByPublicID(tenant.TenantContext, string) (*iamrole.IAMRole, error) {
	if f.roleErr != nil {
		return nil, f.roleErr
	}
	return f.role, nil
}
func (f *fakeRepository) ListAllPermissions() ([]permission.IAMPermission, error) {
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
func (f *fakeRepository) ResolvePermissionSelection([]permission.IAMPermission, []string) ([]string, map[string]bool, []uint, error) {
	if f.selectionErr != nil {
		return nil, nil, nil, f.selectionErr
	}
	return f.normalized, f.selected, f.ids, nil
}
func (f *fakeRepository) BuildRolePermissionCatalog([]permission.IAMPermission, map[string]bool) []iamrole.RolePermissionGroupResponse {
	return []iamrole.RolePermissionGroupResponse{{Module: "roles"}}
}

func TestServiceAssignProtectedTargets(t *testing.T) {
	id := uint(7)
	for _, tt := range []struct {
		name string
		tc   tenant.TenantContext
		role *iamrole.IAMRole
		want error
	}{
		{"own custom", tenant.NewScoped(id, "own"), &iamrole.IAMRole{Slug: "custom"}, nil},
		{"system grants editable", tenant.NewScoped(id, "own"), &iamrole.IAMRole{Slug: "system", IsSystem: true}, nil},
		{"reserved user grants editable", tenant.NewScoped(id, "own"), &iamrole.IAMRole{Name: " User ", Slug: roles.UserRoleSlug, IsSystem: true}, nil},
		{"root custom", tenant.NewRoot(), &iamrole.IAMRole{Slug: "custom"}, nil},
		{"root system grants editable", tenant.NewRoot(), &iamrole.IAMRole{Slug: "system", IsSystem: true}, nil},
		{"root target protected", tenant.NewRoot(), &iamrole.IAMRole{Slug: roles.RootRoleSlug, IsSystem: true}, core.ErrSystemImmutable},
		{"normalized root name protected", tenant.NewScoped(id, "own"), &iamrole.IAMRole{Name: " ROOT ", Slug: "custom"}, core.ErrSystemImmutable},
		{"normalized admin name protected", tenant.NewRoot(), &iamrole.IAMRole{Name: " Admin ", Slug: "custom"}, core.ErrSystemImmutable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.role.ID = 11
			repo := &fakeRepository{role: tt.role, selected: map[string]bool{}, ids: []uint{}}
			_, err := NewService(repo, nil).Assign(tt.tc, "custom", iamrole.AssignRolePermissionsRequest{Permissions: []string{}})
			if !errors.Is(err, tt.want) {
				t.Fatalf("want %v, got %v", tt.want, err)
			}
			if tt.want != nil && repo.replacedRoleID != 0 {
				t.Fatal("protected target persisted permissions")
			}
			if tt.want == nil && (repo.replacedRoleID != 11 || len(repo.replacedPermIDs) != 0) {
				t.Fatal("editable target did not persist empty grants")
			}
		})
	}
}

func TestServiceAssign(t *testing.T) {
	errRepo := errors.New("repo")
	baseRole := &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 11, PublicID: "role-11"}, Slug: "manager"}

	tests := []struct {
		name    string
		repo    *fakeRepository
		wantErr error
	}{
		{name: "custom target does not repeat route permission checks", repo: &fakeRepository{role: baseRole}},
		{name: "role not found", repo: &fakeRepository{roleErr: core.ErrNotFound}, wantErr: core.ErrNotFound},
		{name: "immutable admin role", repo: &fakeRepository{role: &iamrole.IAMRole{Slug: roles.AdminRoleSlug}}, wantErr: core.ErrSystemImmutable},
		{name: "selection error propagated", repo: &fakeRepository{role: baseRole, selectionErr: core.ErrInvalidPermissionSlug}, wantErr: core.ErrInvalidPermissionSlug},
		{name: "can remove all system permissions", repo: &fakeRepository{role: &iamrole.IAMRole{IsSystem: true}, selected: map[string]bool{}}},
		{name: "replace error propagated", repo: &fakeRepository{role: baseRole, selected: map[string]bool{"roles.read": true}, ids: []uint{1}, replaceErr: errRepo}, wantErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil)
			_, err := svc.Assign(tenant.NewRoot(), "role-11", iamrole.AssignRolePermissionsRequest{Permissions: []string{"roles.read"}})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	t.Run("success returns response and writes ids", func(t *testing.T) {
		repo := &fakeRepository{
			role:       baseRole,
			catalog:    []permission.IAMPermission{{Slug: "roles.read"}},
			normalized: []string{"roles.read"},
			selected:   map[string]bool{"roles.read": true},
			ids:        []uint{7},
		}
		svc := NewService(repo, nil)

		tc := tenant.NewScoped(7, "selected-company")
		res, err := svc.Assign(tc, "role-11", iamrole.AssignRolePermissionsRequest{Permissions: []string{"roles.read"}})
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
