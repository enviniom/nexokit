package update_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepository struct {
	role         *iamrole.IAMRole
	getErr       error
	nameExists   bool
	slugExists   bool
	nameErr      error
	slugErr      error
	updateErr    error
	updated      *iamrole.IAMRole
	updatedScope tenant.TenantContext
}

func (f *fakeRepository) GetRoleByPublicID(_ tenant.TenantContext, _ string) (*iamrole.IAMRole, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.role, nil
}

func (f *fakeRepository) ExistsRoleByName(_ tenant.TenantContext, _ string, _ uint) (bool, error) {
	if f.nameErr != nil {
		return false, f.nameErr
	}
	return f.nameExists, nil
}

func (f *fakeRepository) ExistsRoleBySlug(_ tenant.TenantContext, _ string, _ uint) (bool, error) {
	if f.slugErr != nil {
		return false, f.slugErr
	}
	return f.slugExists, nil
}

func (f *fakeRepository) UpdateRole(tc tenant.TenantContext, role *iamrole.IAMRole) error {
	f.updatedScope = tc
	f.updated = role
	return f.updateErr
}

func TestUpdateRoleService(t *testing.T) {
	errName := errors.New("name exists failure")
	errSlug := errors.New("slug exists failure")
	errUpdate := errors.New("update failure")

	tests := []struct {
		name         string
		repo         *fakeRepository
		req          iamrole.UpdateRoleRequest
		wantErr      error
		assertUpdate bool
	}{
		{name: "success", repo: &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-1"}, Name: "Manager", Slug: "manager", Description: "old"}}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead", Description: "new"}, assertUpdate: true},
		{name: "not found propagation", repo: &fakeRepository{getErr: core.ErrNotFound}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: core.ErrNotFound},
		{name: "protected existing role by system", repo: &fakeRepository{role: &iamrole.IAMRole{IsSystem: true, Name: "Custom", Slug: "custom"}}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: core.ErrRoleProtected},
		{name: "protected existing role by reserved identity", repo: &fakeRepository{role: &iamrole.IAMRole{Name: "Root", Slug: "custom"}}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: core.ErrRoleProtected},
		{name: "reserved request identity", repo: &fakeRepository{role: &iamrole.IAMRole{Name: "Custom", Slug: "custom"}}, req: iamrole.UpdateRoleRequest{Name: "Admin", Slug: "lead"}, wantErr: core.ErrReservedRoleIdentity},
		{name: "duplicate name", repo: &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 7}, Name: "Custom", Slug: "custom"}, nameExists: true}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: core.ErrRoleNameAlreadyExists},
		{name: "duplicate slug", repo: &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 7}, Name: "Custom", Slug: "custom"}, slugExists: true}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: core.ErrRoleSlugAlreadyExists},
		{name: "exists name error", repo: &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 7}, Name: "Custom", Slug: "custom"}, nameErr: errName}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: errName},
		{name: "exists slug error", repo: &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 7}, Name: "Custom", Slug: "custom"}, slugErr: errSlug}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: errSlug},
		{name: "update error", repo: &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-1"}, Name: "Custom", Slug: "custom"}, updateErr: errUpdate}, req: iamrole.UpdateRoleRequest{Name: "Lead", Slug: "lead"}, wantErr: errUpdate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			_, err := svc.Update(tenant.NewRoot(), "role-1", tt.req)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.assertUpdate {
					if tt.repo.updated == nil {
						t.Fatalf("expected update to be called")
					}
					if tt.repo.updated.Name != tt.req.Name || tt.repo.updated.Slug != tt.req.Slug {
						t.Fatalf("expected updated role with request values")
					}
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
