package delete_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeServiceRepository struct {
	role      *core.IAMRole
	getErr    error
	count     int64
	countErr  error
	deleteErr error

	deletedPublicID string
}

func (f *fakeServiceRepository) GetByPublicID(_ tenant.TenantContext, _ string) (*core.IAMRole, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.role, nil
}

func (f *fakeServiceRepository) CountUsersByRoleID(_ uint) (int64, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	return f.count, nil
}

func (f *fakeServiceRepository) DeleteByPublicID(_ tenant.TenantContext, publicID string) error {
	f.deletedPublicID = publicID
	return f.deleteErr
}

func TestDeleteRoleServiceDelete(t *testing.T) {
	errCount := errors.New("count error")
	errDelete := errors.New("delete error")

	tests := []struct {
		name           string
		repo           *fakeServiceRepository
		expectedErr    error
		expectDeleteID string
	}{
		{
			name: "success",
			repo: &fakeServiceRepository{role: &core.IAMRole{BaseModel: shared.BaseModel{ID: 11}, Name: "Custom", Slug: "custom"}},
		},
		{
			name:        "not found propagation",
			repo:        &fakeServiceRepository{getErr: core.ErrNotFound},
			expectedErr: core.ErrNotFound,
		},
		{
			name:        "protected role by system flag",
			repo:        &fakeServiceRepository{role: &core.IAMRole{Name: "Custom", Slug: "custom", IsSystem: true}},
			expectedErr: core.ErrRoleProtected,
		},
		{
			name:        "reserved identity",
			repo:        &fakeServiceRepository{role: &core.IAMRole{Name: "Root", Slug: "custom"}},
			expectedErr: core.ErrRoleProtected,
		},
		{
			name:        "assigned users",
			repo:        &fakeServiceRepository{role: &core.IAMRole{BaseModel: shared.BaseModel{ID: 4}, Name: "Custom", Slug: "custom"}, count: 1},
			expectedErr: core.ErrRoleHasAssignedUsers,
		},
		{
			name:        "count users repo error",
			repo:        &fakeServiceRepository{role: &core.IAMRole{BaseModel: shared.BaseModel{ID: 4}, Name: "Custom", Slug: "custom"}, countErr: errCount},
			expectedErr: errCount,
		},
		{
			name:        "delete repo error",
			repo:        &fakeServiceRepository{role: &core.IAMRole{BaseModel: shared.BaseModel{ID: 4}, Name: "Custom", Slug: "custom"}, deleteErr: errDelete},
			expectedErr: errDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			err := svc.Delete(tenant.NewRoot(), "role-1")
			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if tt.repo.deletedPublicID != "role-1" {
					t.Fatalf("expected delete call with role-1, got %q", tt.repo.deletedPublicID)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error")
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
