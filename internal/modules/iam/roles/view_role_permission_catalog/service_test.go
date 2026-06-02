package view_role_permission_catalog

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeServiceRepo struct {
	role    *core.IAMRole
	catalog []core.IAMPermission
	roleErr error
	catErr  error
}

func (f fakeServiceRepo) GetRoleByPublicID(tenant.TenantContext, string) (*core.IAMRole, error) {
	if f.roleErr != nil {
		return nil, f.roleErr
	}
	return f.role, nil
}

func (f fakeServiceRepo) ListPermissionCatalog() ([]core.IAMPermission, error) {
	if f.catErr != nil {
		return nil, f.catErr
	}
	return f.catalog, nil
}

func TestServiceViewSuccess(t *testing.T) {
	svc := NewService(fakeServiceRepo{
		role: &core.IAMRole{Permissions: []core.IAMPermission{{Slug: "users.read"}}},
		catalog: []core.IAMPermission{
			{BaseModel: shared.BaseModel{PublicID: "perm-1"}, Slug: "users.read", Name: "Read Users", Module: "users", Action: "read", DisplayOrder: 1},
			{BaseModel: shared.BaseModel{PublicID: "perm-2"}, Slug: "users.write", Name: "Write Users", Module: "users", Action: "write", DisplayOrder: 2},
			{BaseModel: shared.BaseModel{PublicID: "perm-3"}, Slug: "roles.read", Name: "Read Roles", Module: "roles", Action: "read", DisplayOrder: 1},
		},
	})

	groups, err := svc.View(tenant.NewRoot(), "role-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 module groups, got %d", len(groups))
	}
	if !groups[0].Permissions[0].Granted {
		t.Fatalf("expected users.read to be granted")
	}
	if groups[0].Permissions[1].Granted {
		t.Fatalf("expected users.write to be not granted")
	}
}

func TestServiceViewRoleErrorPropagation(t *testing.T) {
	svc := NewService(fakeServiceRepo{roleErr: core.ErrNotFound})

	_, err := svc.View(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceViewCatalogErrorPropagation(t *testing.T) {
	repoErr := errors.New("permission catalog unavailable")
	svc := NewService(fakeServiceRepo{role: &core.IAMRole{}, catErr: repoErr})

	_, err := svc.View(tenant.NewRoot(), "role-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
