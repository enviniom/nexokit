package create_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct {
	rootRole    *core.IAMRole
	roleSlugErr error

	emailExists bool
	emailErr    error

	createErr error

	created    *core.UserResponse
	getByIDErr error
}

func (f fakeRepo) GetRoleBySlug(string) (*core.IAMRole, error) {
	return f.rootRole, f.roleSlugErr
}

func (f fakeRepo) ExistsByEmail(string) (bool, error) {
	return f.emailExists, f.emailErr
}

func (f fakeRepo) Create(*core.IAMUser) error {
	return f.createErr
}

func (f fakeRepo) GetByPublicID(tenant.TenantContext, string) (*core.UserResponse, error) {
	return f.created, f.getByIDErr
}

type fakeHasher struct {
	hash string
	err  error
}

func (f fakeHasher) HashPassword(string) (string, error) {
	return f.hash, f.err
}

func rootRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 1, PublicID: "role-root"}, Slug: "root"}
}

func userRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 2, PublicID: "role-user"}, Slug: "user"}
}

func TestCreateService(t *testing.T) {
	companyID := uint(10)

	tests := []struct {
		name    string
		repo    fakeRepo
		hasher  fakeHasher
		tc      tenant.TenantContext
		req     core.CreateUserRequest
		wantErr error
	}{
		{
			name: "success root scope with root role",
			repo: fakeRepo{
				rootRole: rootRole(),
				created:  &core.UserResponse{PublicID: "user-1", Name: "Root User"},
			},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "Root User", Email: "root@example.com", Password: "password123", RoleID: 1},
		},
		{
			name: "success scoped with user role",
			repo: fakeRepo{
				rootRole: rootRole(),
				created:  &core.UserResponse{PublicID: "user-2", Name: "Alice", CompanyID: &companyID},
			},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewScoped(companyID, "acme"),
			req:    core.CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "password123", RoleID: 2},
		},
		{
			name:    "role slug lookup fails",
			repo:    fakeRepo{roleSlugErr: errors.New("db down")},
			hasher:  fakeHasher{hash: "hashed"},
			tc:      tenant.NewRoot(),
			req:     core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1},
			wantErr: errors.New("db down"),
		},
		{
			name: "root role from non-root scope forbidden",
			repo: fakeRepo{rootRole: rootRole()},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewScoped(companyID, "acme"),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1},
			wantErr: core.ErrForbiddenRoleAssignment,
		},
		{
			name: "root role with company_id forbidden",
			repo: fakeRepo{rootRole: rootRole()},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1, CompanyID: &companyID},
			wantErr: core.ErrForbiddenRoleAssignment,
		},
		{
			name: "root scope without company_id invalid",
			repo: fakeRepo{rootRole: rootRole()},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 2},
			wantErr: core.ErrInvalidCompanyScope,
		},
		{
			name: "cross-company assignment forbidden",
			repo: fakeRepo{rootRole: rootRole()},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewScoped(companyID, "acme"),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 2, CompanyID: ptrUint(99)},
			wantErr: core.ErrForbiddenRoleAssignment,
		},
		{
			name: "duplicate email",
			repo: fakeRepo{rootRole: rootRole(), emailExists: true},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "taken@example.com", Password: "password123", RoleID: 1},
			wantErr: core.ErrUserEmailAlreadyExists,
		},
		{
			name: "email check error",
			repo: fakeRepo{rootRole: rootRole(), emailErr: errors.New("db down")},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1},
			wantErr: errors.New("db down"),
		},
		{
			name: "hash error",
			repo: fakeRepo{rootRole: rootRole()},
			hasher: fakeHasher{err: errors.New("hash failed")},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1},
			wantErr: errors.New("hash failed"),
		},
		{
			name: "create error",
			repo: fakeRepo{rootRole: rootRole(), createErr: errors.New("insert failed")},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1},
			wantErr: errors.New("insert failed"),
		},
		{
			name: "get by public id error",
			repo: fakeRepo{rootRole: rootRole(), getByIDErr: errors.New("reload failed")},
			hasher: fakeHasher{hash: "hashed"},
			tc:     tenant.NewRoot(),
			req:    core.CreateUserRequest{Name: "X", Email: "x@example.com", Password: "password123", RoleID: 1},
			wantErr: errors.New("reload failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, tt.hasher)
			resp, err := svc.Create(tt.tc, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatalf("expected response, got nil")
			}
		})
	}
}

func ptrUint(v uint) *uint { return &v }
