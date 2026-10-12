package create_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeRepo struct {
	nameExists bool
	slugExists bool
	createErr  error
	errName    error
	errSlug    error
	item       *iamrole.IAMRole
}

func (f fakeRepo) ExistsRoleByName(_ tenant.TenantContext, _ string) (bool, error) {
	if f.errName != nil {
		return false, f.errName
	}
	return f.nameExists, nil
}

func (f fakeRepo) ExistsRoleBySlug(_ tenant.TenantContext, _ string) (bool, error) {
	if f.errSlug != nil {
		return false, f.errSlug
	}
	return f.slugExists, nil
}

func (f fakeRepo) Create(_ tenant.TenantContext, _ iamrole.CreateRoleRequest) (*iamrole.IAMRole, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.item != nil {
		return f.item, nil
	}
	return &iamrole.IAMRole{}, nil
}

func TestCreateRoleService(t *testing.T) {
	tests := []struct {
		name    string
		repo    fakeRepo
		req     iamrole.CreateRoleRequest
		wantErr error
	}{
		{name: "success", repo: fakeRepo{item: &iamrole.IAMRole{Name: "Manager", Slug: "manager"}}, req: iamrole.CreateRoleRequest{Name: "Manager", Slug: "manager"}},
		{name: "reserved identity", repo: fakeRepo{}, req: iamrole.CreateRoleRequest{Name: "Root", Slug: "manager"}, wantErr: core.ErrReservedRoleIdentity},
		{name: "duplicate name", repo: fakeRepo{nameExists: true}, req: iamrole.CreateRoleRequest{Name: "Manager", Slug: "manager"}, wantErr: core.ErrRoleNameAlreadyExists},
		{name: "duplicate slug", repo: fakeRepo{slugExists: true}, req: iamrole.CreateRoleRequest{Name: "Manager", Slug: "manager"}, wantErr: core.ErrRoleSlugAlreadyExists},
		{name: "repo exists error", repo: fakeRepo{errName: errors.New("db down")}, req: iamrole.CreateRoleRequest{Name: "Manager", Slug: "manager"}, wantErr: errors.New("db down")},
		{name: "repo create error", repo: fakeRepo{createErr: errors.New("insert failed")}, req: iamrole.CreateRoleRequest{Name: "Manager", Slug: "manager"}, wantErr: errors.New("insert failed")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			_, err := svc.Create(tenant.NewRoot(), tt.req)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
