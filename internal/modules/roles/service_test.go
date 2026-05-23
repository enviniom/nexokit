package roles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// fakeRepository is a test double for the repository.
type fakeRepository struct {
	roles                 []Role
	roleByName            map[string]*Role
	total                 int64
	err                   error
	getByNameErr          error
	createErr             error
	updateErr             error
	assignedRoleID        uint
	assignedPermissionIDs []uint
}

func roleMatchesTenantScope(role Role, tc tenant.TenantContext) bool {
	if tc.IsRootScope {
		return true
	}
	if role.CompanyID == nil {
		return false
	}
	return *role.CompanyID == tc.CompanyID
}

func (f *fakeRepository) List(tc tenant.TenantContext, page, perPage int) ([]Role, error) {
	if f.err != nil {
		return nil, f.err
	}

	if tc.IsRootScope {
		return f.roles, nil
	}

	items := make([]Role, 0, len(f.roles))
	for _, role := range f.roles {
		if roleMatchesTenantScope(role, tc) {
			items = append(items, role)
		}
	}
	return items, nil
}

func (f *fakeRepository) ListSelect(tc tenant.TenantContext) ([]Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	var selectRes []Role
	for _, r := range f.roles {
		if roleMatchesTenantScope(r, tc) && r.Slug != RootRoleSlug {
			selectRes = append(selectRes, r)
		}
	}
	return selectRes, nil
}

func (f *fakeRepository) Count(tc tenant.TenantContext) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	if tc.IsRootScope {
		if f.total > 0 {
			return f.total, nil
		}
		return int64(len(f.roles)), nil
	}

	var count int64
	for _, role := range f.roles {
		if roleMatchesTenantScope(role, tc) {
			count++
		}
	}
	if f.total > 0 {
		return f.total, nil
	}
	return count, nil
}

func (f *fakeRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.roles {
		if f.roles[i].PublicID == publicID {
			if !roleMatchesTenantScope(f.roles[i], tc) {
				return nil, gorm.ErrRecordNotFound
			}
			return &f.roles[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetByName(tc tenant.TenantContext, name string) (*Role, error) {
	if f.getByNameErr != nil {
		return nil, f.getByNameErr
	}
	if f.err != nil {
		return nil, f.err
	}
	if r, ok := f.roleByName[name]; ok {
		if !roleMatchesTenantScope(*r, tc) {
			return nil, gorm.ErrRecordNotFound
		}
		return r, nil
	}
	for i := range f.roles {
		if f.roles[i].Name == name {
			if !roleMatchesTenantScope(f.roles[i], tc) {
				return nil, gorm.ErrRecordNotFound
			}
			return &f.roles[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetBySlug(tc tenant.TenantContext, slug string) (*Role, error) {
	if f.getByNameErr != nil {
		return nil, f.getByNameErr
	}
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.roles {
		if f.roles[i].Slug == slug {
			if !roleMatchesTenantScope(f.roles[i], tc) {
				return nil, gorm.ErrRecordNotFound
			}
			return &f.roles[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) Create(role *Role) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.err != nil {
		return f.err
	}
	f.roles = append(f.roles, *role)
	return nil
}

func (f *fakeRepository) Update(role *Role) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if f.err != nil {
		return f.err
	}
	for i := range f.roles {
		if f.roles[i].PublicID == role.PublicID {
			f.roles[i] = *role
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeRepository) Delete(tc tenant.TenantContext, publicID string) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.roles {
		if f.roles[i].PublicID == publicID {
			if !roleMatchesTenantScope(f.roles[i], tc) {
				return gorm.ErrRecordNotFound
			}
			f.roles = append(f.roles[:i], f.roles[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeRepository) ReplacePermissions(roleID uint, permissionIDs []uint) error {
	f.assignedRoleID = roleID
	f.assignedPermissionIDs = append([]uint(nil), permissionIDs...)
	for i := range f.roles {
		if f.roles[i].ID == roleID {
			f.roles[i].Permissions = nil
			for _, id := range permissionIDs {
				f.roles[i].Permissions = append(f.roles[i].Permissions, permissions.Permission{BaseModel: shared.BaseModel{ID: id}})
			}
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeRepository) ExistsByName(tc tenant.TenantContext, name string) (bool, error) {
	_, err := f.GetByName(tc, name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (f *fakeRepository) ExistsBySlug(tc tenant.TenantContext, slug string) (bool, error) {
	_, err := f.GetBySlug(tc, slug)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

type fakePermissionCatalogRepository struct {
	permissions []permissions.Permission
	err         error
}

func (f fakePermissionCatalogRepository) ListAll() ([]permissions.Permission, error) {
	return f.permissions, f.err
}

type fakeRoleMemberRepository struct {
	publicIDs []string
	count     int64
	err       error
}

func (f fakeRoleMemberRepository) ListPublicIDsByRoleID(roleID uint) ([]string, error) {
	return f.publicIDs, f.err
}

func (f fakeRoleMemberRepository) CountByRoleID(roleID uint) (int64, error) {
	return f.count, f.err
}

type fakeCache struct {
	deleted []string
}

func (f *fakeCache) Get(ctx context.Context, key string) ([]byte, error) { return nil, nil }
func (f *fakeCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}
func (f *fakeCache) Delete(ctx context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}
func (f *fakeCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (f *fakeCache) Close() error                                         { return nil }

func TestService_List(t *testing.T) {
	t.Run("returns paginated roles", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "admin", IsSystem: true},
				{BaseModel: shared.BaseModel{PublicID: "role2"}, Name: "user", IsSystem: true},
			},
			total: 2,
		}
		svc := NewService(repo)

		result, total, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 roles, got %d", len(result))
		}
		if result[0].Name != "admin" {
			t.Errorf("expected first role name 'admin', got %s", result[0].Name)
		}
		if !result[0].IsSystem {
			t.Error("expected first role to be system")
		}
	})

	t.Run("returns empty list when no roles", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{}, total: 0}
		svc := NewService(repo)

		result, total, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 roles, got %d", len(result))
		}
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		repo := &fakeRepository{err: apperror.ErrInternal}
		svc := NewService(repo)

		_, _, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestService_GetByPublicID(t *testing.T) {
	t.Run("returns role when found", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "admin", IsSystem: true},
			},
		}
		svc := NewService(repo)

		result, err := svc.GetByPublicID(tenant.NewRoot(), "role1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.PublicID != "role1" {
			t.Errorf("expected public_id 'role1', got %s", result.PublicID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{}}
		svc := NewService(repo)

		_, err := svc.GetByPublicID(tenant.NewRoot(), "missing")
		if err == nil {
			t.Error("expected error for missing role")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestService_Create(t *testing.T) {
	t.Run("creates a new role successfully", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{}}
		svc := NewService(repo)

		req := CreateRoleRequest{Name: "editor", Slug: "editor", Description: "Can edit content"}
		result, err := svc.Create(tenant.NewRoot(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "editor" {
			t.Errorf("expected name 'editor', got %s", result.Name)
		}
		if result.Slug != "editor" {
			t.Errorf("expected slug 'editor', got %s", result.Slug)
		}
	})

	t.Run("returns conflict when name already exists", func(t *testing.T) {
		repo := &fakeRepository{
			roleByName: map[string]*Role{
				"editor": {BaseModel: shared.BaseModel{PublicID: "r1"}, Name: "editor"},
			},
		}
		svc := NewService(repo)

		req := CreateRoleRequest{Name: "editor", Slug: "editor"}
		_, err := svc.Create(tenant.NewRoot(), req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns conflict when slug already exists", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{{BaseModel: shared.BaseModel{PublicID: "r1"}, Name: "admin", Slug: "editor"}}}
		svc := NewService(repo)

		req := CreateRoleRequest{Name: "manager", Slug: "editor"}
		_, err := svc.Create(tenant.NewRoot(), req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns validation when creating reserved role identity", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{}}
		svc := NewService(repo)

		tests := []CreateRoleRequest{
			{Name: "root", Slug: "custom-role"},
			{Name: "Root", Slug: "custom-role"},
			{Name: "custom", Slug: "root"},
			{Name: "admin", Slug: "custom-role"},
			{Name: "custom", Slug: "user"},
		}

		for _, req := range tests {
			_, err := svc.Create(tenant.NewRoot(), req)
			if !errors.Is(err, apperror.ErrValidation) {
				t.Fatalf("expected ErrValidation for %+v, got %v", req, err)
			}
		}
	})

	t.Run("returns repository error when uniqueness check fails", func(t *testing.T) {
		repo := &fakeRepository{getByNameErr: apperror.ErrInternal}
		svc := NewService(repo)

		req := CreateRoleRequest{Name: "editor", Slug: "editor"}
		_, err := svc.Create(tenant.NewRoot(), req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrInternal) {
			t.Errorf("expected ErrInternal, got %v", err)
		}
	})

	t.Run("returns conflict when repository create hits unique constraint", func(t *testing.T) {
		repo := &fakeRepository{createErr: gorm.ErrDuplicatedKey}
		svc := NewService(repo)

		req := CreateRoleRequest{Name: "editor", Slug: "editor"}
		_, err := svc.Create(tenant.NewRoot(), req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestService_Update(t *testing.T) {
	t.Run("updates a non-system role successfully", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", IsSystem: false},
			},
			roleByName: map[string]*Role{},
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "senior-editor", Slug: "senior-editor"}
		result, err := svc.Update(tenant.NewRoot(), "role1", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "senior-editor" {
			t.Errorf("expected name 'senior-editor', got %s", result.Name)
		}
	})

	t.Run("returns forbidden when updating a system role", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "root", IsSystem: true},
			},
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "super-root"}
		_, err := svc.Update(tenant.NewRoot(), "role1", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns validation when updating reserved role identity", func(t *testing.T) {
		tests := []struct {
			role Role
			req  UpdateRoleRequest
		}{
			{role: Role{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "root", Slug: "custom-role", IsSystem: false}, req: UpdateRoleRequest{Name: "renamed-root", Slug: "renamed-root"}},
			{role: Role{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "custom-role", Slug: RootRoleSlug, IsSystem: false}, req: UpdateRoleRequest{Name: "renamed-root", Slug: "renamed-root"}},
			{role: Role{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "manager", Slug: "manager", IsSystem: false}, req: UpdateRoleRequest{Name: "admin", Slug: "manager"}},
			{role: Role{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "manager", Slug: "manager", IsSystem: false}, req: UpdateRoleRequest{Name: "manager", Slug: "user"}},
		}

		for _, tt := range tests {
			repo := &fakeRepository{roles: []Role{tt.role}}
			svc := NewService(repo)
			_, err := svc.Update(tenant.NewRoot(), "role1", tt.req)
			if !errors.Is(err, apperror.ErrValidation) {
				t.Fatalf("expected ErrValidation for role %+v request %+v, got %v", tt.role, tt.req, err)
			}
		}
	})

	t.Run("returns validation for requested reserved identity before lookup", func(t *testing.T) {
		repo := &fakeRepository{err: apperror.ErrInternal}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "admin", Slug: "admin"}
		_, err := svc.Update(tenant.NewRoot(), "missing-role", req)
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation before repository lookup, got %v", err)
		}
	})

	t.Run("returns conflict when updating to an existing name", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", IsSystem: false},
				{BaseModel: shared.BaseModel{PublicID: "role2"}, Name: "manager", IsSystem: false},
			},
			roleByName: map[string]*Role{
				"manager": {BaseModel: shared.BaseModel{PublicID: "role2"}, Name: "manager"},
			},
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "manager"}
		_, err := svc.Update(tenant.NewRoot(), "role1", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns conflict when updating slug to one used by another role", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", Slug: "editor", IsSystem: false},
				{BaseModel: shared.BaseModel{PublicID: "role2"}, Name: "manager", Slug: "manager", IsSystem: false},
			},
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "editor", Slug: "manager"}
		_, err := svc.Update(tenant.NewRoot(), "role1", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns repository error when uniqueness check fails", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", IsSystem: false},
			},
			getByNameErr: apperror.ErrInternal,
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "manager"}
		_, err := svc.Update(tenant.NewRoot(), "role1", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrInternal) {
			t.Errorf("expected ErrInternal, got %v", err)
		}
	})

	t.Run("returns conflict when repository update hits unique constraint", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", IsSystem: false},
			},
			roleByName: map[string]*Role{},
			updateErr:  gorm.ErrDuplicatedKey,
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "manager"}
		_, err := svc.Update(tenant.NewRoot(), "role1", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("deletes a non-system role successfully", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", IsSystem: false},
			},
		}
		svc := NewService(repo)

		err := svc.Delete(tenant.NewRoot(), "role1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns forbidden when deleting a system role", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "root", IsSystem: true},
			},
		}
		svc := NewService(repo)

		err := svc.Delete(tenant.NewRoot(), "role1")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns forbidden when deleting root role even if not system", func(t *testing.T) {
		tests := []Role{
			{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "root", Slug: "custom-role", IsSystem: false},
			{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: " Root ", Slug: "custom-role", IsSystem: false},
			{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "custom-role", Slug: RootRoleSlug, IsSystem: false},
			{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "custom-role", Slug: " ROOT ", IsSystem: false},
		}

		for _, role := range tests {
			repo := &fakeRepository{roles: []Role{role}}
			svc := NewService(repo)

			err := svc.Delete(tenant.NewRoot(), "role1")
			if !errors.Is(err, apperror.ErrForbidden) {
				t.Fatalf("expected ErrForbidden for role %+v, got %v", role, err)
			}
		}
	})

	t.Run("returns unprocessable when role has assigned users", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{{BaseModel: shared.BaseModel{ID: 9, PublicID: "role1"}, Name: "editor", IsSystem: false}}}
		svc := NewService(repo, WithRoleMembers(fakeRoleMemberRepository{count: 2}))

		err := svc.Delete(tenant.NewRoot(), "role1")
		if !errors.Is(err, apperror.ErrUnprocessable) {
			t.Fatalf("expected ErrUnprocessable, got %v", err)
		}
	})
}

func TestService_GetPermissionCatalog(t *testing.T) {
	t.Run("returns grouped catalog with granted flags", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-admin"}, Name: "admin", Permissions: []permissions.Permission{{Slug: "users.list"}}}}}
		catalog := fakePermissionCatalogRepository{permissions: []permissions.Permission{
			{BaseModel: shared.BaseModel{PublicID: "p1"}, Module: "roles", Slug: "roles.list", Action: permissions.ActionList, DisplayOrder: 10},
			{BaseModel: shared.BaseModel{PublicID: "p2"}, Module: "users", Slug: "users.list", Action: permissions.ActionList, DisplayOrder: 10},
			{BaseModel: shared.BaseModel{PublicID: "p3"}, Module: "users", Slug: "users.view", Action: permissions.ActionView, DisplayOrder: 20},
		}}
		svc := NewService(repo, WithPermissionCatalog(catalog))

		groups, err := svc.GetPermissionCatalog(tenant.NewRoot(), "role-admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 2 || groups[0].Module != "roles" || groups[1].Module != "users" {
			t.Fatalf("expected roles and users groups, got %+v", groups)
		}
		if !groups[1].Permissions[0].Granted || groups[1].Permissions[1].Granted {
			t.Fatalf("expected granted flag only for users.list, got %+v", groups[1].Permissions)
		}
	})

	t.Run("returns not found for missing role", func(t *testing.T) {
		svc := NewService(&fakeRepository{roles: []Role{}}, WithPermissionCatalog(fakePermissionCatalogRepository{}))
		_, err := svc.GetPermissionCatalog(tenant.NewRoot(), "missing")
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})
}

func TestService_AssignPermissions(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		catalog    []permissions.Permission
		req        AssignRolePermissionsRequest
		actorPerms []string
		wantErr    error
		wantIDs    []uint
		wantDelete []string
	}{
		{
			name:       "replaces role permissions exactly and invalidates member cache",
			role:       Role{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-admin"}, Name: "admin", Permissions: []permissions.Permission{{BaseModel: shared.BaseModel{ID: 1}, Slug: "users.list"}}},
			catalog:    []permissions.Permission{{BaseModel: shared.BaseModel{ID: 2}, Slug: "users.view", Module: "users"}, {BaseModel: shared.BaseModel{ID: 3}, Slug: "roles.list", Module: "roles"}},
			req:        AssignRolePermissionsRequest{Permissions: []string{"users.view", "roles.list"}},
			actorPerms: []string{"roles.assign_permissions"},
			wantIDs:    []uint{2, 3},
			wantDelete: []string{"rbac:permissions:user-one", "rbac:permissions:user-two"},
		},
		{
			name:       "rejects invalid slug",
			role:       Role{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-admin"}, Name: "admin"},
			catalog:    []permissions.Permission{{BaseModel: shared.BaseModel{ID: 2}, Slug: "users.view", Module: "users"}},
			req:        AssignRolePermissionsRequest{Permissions: []string{"missing.slug"}},
			actorPerms: []string{"roles.assign_permissions"},
			wantErr:    apperror.ErrBadRequest,
		},
		{
			name:       "rejects unauthorized assignment",
			role:       Role{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-admin"}, Name: "admin"},
			catalog:    []permissions.Permission{{BaseModel: shared.BaseModel{ID: 2}, Slug: "users.view", Module: "users"}},
			req:        AssignRolePermissionsRequest{Permissions: []string{"users.view"}},
			actorPerms: []string{"roles.view"},
			wantErr:    apperror.ErrForbidden,
		},
		{
			name:       "protects system role system permissions",
			role:       Role{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-root"}, Name: "root", IsSystem: true, Permissions: []permissions.Permission{{BaseModel: shared.BaseModel{ID: 1}, Slug: "users.list", IsSystem: true}}},
			catalog:    []permissions.Permission{{BaseModel: shared.BaseModel{ID: 1}, Slug: "users.list", Module: "users", IsSystem: true}, {BaseModel: shared.BaseModel{ID: 2}, Slug: "users.view", Module: "users"}},
			req:        AssignRolePermissionsRequest{Permissions: []string{"users.view"}},
			actorPerms: []string{"roles.assign_permissions"},
			wantErr:    apperror.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{roles: []Role{tt.role}}
			cache := &fakeCache{}
			svc := NewService(repo, WithPermissionCatalog(fakePermissionCatalogRepository{permissions: tt.catalog}), WithRoleMembers(fakeRoleMemberRepository{publicIDs: []string{"user-one", "user-two"}}), WithCache(cache))

			_, err := svc.AssignPermissions(tenant.NewRoot(), tt.role.PublicID, tt.req, tt.actorPerms)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				if len(repo.assignedPermissionIDs) != 0 {
					t.Fatalf("expected no replacement on error, got %v", repo.assignedPermissionIDs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !uintSlicesEqual(repo.assignedPermissionIDs, tt.wantIDs) {
				t.Fatalf("expected assigned IDs %v, got %v", tt.wantIDs, repo.assignedPermissionIDs)
			}
			if !stringSlicesEqual(cache.deleted, tt.wantDelete) {
				t.Fatalf("expected deleted keys %v, got %v", tt.wantDelete, cache.deleted)
			}
		})
	}
}

func TestService_TenantScopedBehavior(t *testing.T) {
	companyA := uint(10)
	companyB := uint(20)

	repo := &fakeRepository{roles: []Role{
		{BaseModel: shared.BaseModel{PublicID: "global-root"}, Name: "root", Slug: "root", IsSystem: true},
		{BaseModel: shared.BaseModel{PublicID: "a-role"}, Name: "manager-a", Slug: "manager", CompanyID: &companyA},
		{BaseModel: shared.BaseModel{PublicID: "b-role"}, Name: "manager-b", Slug: "manager", CompanyID: &companyB},
	}}
	svc := NewService(repo)

	t.Run("list returns only current company roles for non-root", func(t *testing.T) {
		items, total, err := svc.List(tenant.NewScoped(companyA, "company-a"), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 || len(items) != 1 || items[0].PublicID != "a-role" {
			t.Fatalf("expected only company A role, got total=%d items=%+v", total, items)
		}
	})

	t.Run("list returns all roles for root scope", func(t *testing.T) {
		items, total, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 || len(items) != 3 {
			t.Fatalf("expected 3 roles for root scope, got total=%d len=%d", total, len(items))
		}
	})

	t.Run("cross-tenant get returns not found", func(t *testing.T) {
		_, err := svc.GetByPublicID(tenant.NewScoped(companyA, "company-a"), "b-role")
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("cross-tenant update returns not found", func(t *testing.T) {
		_, err := svc.Update(tenant.NewScoped(companyA, "company-a"), "b-role", UpdateRoleRequest{Name: "new", Slug: "new"})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("cross-tenant delete returns not found", func(t *testing.T) {
		err := svc.Delete(tenant.NewScoped(companyA, "company-a"), "b-role")
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})
}

func TestService_CreateSetsCompanyIDByScope(t *testing.T) {
	companyID := uint(77)
	repo := &fakeRepository{roles: []Role{}}
	svc := NewService(repo)

	t.Run("scoped create sets company id", func(t *testing.T) {
		res, err := svc.Create(tenant.NewScoped(companyID, "acme"), CreateRoleRequest{Name: "editor", Slug: "editor"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.roles) != 1 || repo.roles[0].CompanyID == nil || *repo.roles[0].CompanyID != companyID {
			t.Fatalf("expected persisted company id %d, got roles %+v", companyID, repo.roles)
		}
		if res.CompanyID != nil {
			t.Fatalf("expected nil response company id without preloaded company, got %+v", res.CompanyID)
		}
	})

	t.Run("root create remains global", func(t *testing.T) {
		res, err := svc.Create(tenant.NewRoot(), CreateRoleRequest{Name: "auditor", Slug: "auditor"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.CompanyID != nil {
			t.Fatalf("expected nil company id for root scope, got %+v", res.CompanyID)
		}
	})
}

func TestRoleResponseUsesCompanyPublicID(t *testing.T) {
	companyID := uint(77)
	companyPublicID := "company_public_77"
	repo := &fakeRepository{roles: []Role{
		{
			BaseModel: shared.BaseModel{PublicID: "role-company"},
			Name:      "editor",
			Slug:      "editor",
			CompanyID: &companyID,
			Company:   &companies.Company{BaseModel: shared.BaseModel{ID: companyID, PublicID: companyPublicID}, Slug: "acme"},
		},
		{BaseModel: shared.BaseModel{PublicID: "role-global"}, Name: "auditor", Slug: "auditor"},
	}}
	svc := NewService(repo)

	t.Run("scoped role returns company public id", func(t *testing.T) {
		res, err := svc.GetByPublicID(tenant.NewScoped(companyID, "acme"), "role-company")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.CompanyID == nil || *res.CompanyID != companyPublicID {
			t.Fatalf("expected company public id %q, got %+v", companyPublicID, res.CompanyID)
		}
	})

	t.Run("global role omits company id", func(t *testing.T) {
		res, err := svc.GetByPublicID(tenant.NewRoot(), "role-global")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.CompanyID != nil {
			t.Fatalf("expected nil company id for global role, got %+v", res.CompanyID)
		}
	})
}

func TestIsReservedIdentity(t *testing.T) {
	cases := []struct {
		name string
		slug string
		want bool
	}{
		{name: "Root", slug: "team", want: true},
		{name: "lead", slug: "ADMIN", want: true},
		{name: " manager ", slug: " user ", want: true},
		{name: "custom", slug: "custom", want: false},
	}

	for _, tc := range cases {
		if got := isReservedIdentity(tc.name, tc.slug); got != tc.want {
			t.Fatalf("isReservedIdentity(%q,%q) = %v, want %v", tc.name, tc.slug, got, tc.want)
		}
	}
}

func uintSlicesEqual(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestService_ListSelect(t *testing.T) {
	companyA := uint(10)
	companyB := uint(20)

	repo := &fakeRepository{roles: []Role{
		{BaseModel: shared.BaseModel{PublicID: "global-root"}, Name: "root", Slug: "root", IsSystem: true},
		{BaseModel: shared.BaseModel{PublicID: "a-role"}, Name: "manager-a", Slug: "manager", CompanyID: &companyA, Company: &companies.Company{BaseModel: shared.BaseModel{PublicID: "comp-a"}}},
		{BaseModel: shared.BaseModel{PublicID: "b-role"}, Name: "manager-b", Slug: "manager", CompanyID: &companyB, Company: &companies.Company{BaseModel: shared.BaseModel{PublicID: "comp-b"}}},
		{BaseModel: shared.BaseModel{PublicID: "global-admin"}, Name: "admin", Slug: "admin", IsSystem: true},
	}}
	svc := NewService(repo)

	t.Run("excludes root slug and filters by tenant scope", func(t *testing.T) {
		res, err := svc.ListSelect(tenant.NewScoped(companyA, "company-a"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected exactly 1 role, got %d: %+v", len(res), res)
		}
		if res[0].ID != "a-role" {
			t.Errorf("expected public_id 'a-role', got %s", res[0].ID)
		}
		if res[0].Meta["company_id"] != "comp-a" {
			t.Errorf("expected company_id 'comp-a' in meta, got %v", res[0].Meta["company_id"])
		}
	})

	t.Run("excludes root slug but returns other roles in root scope", func(t *testing.T) {
		res, err := svc.ListSelect(tenant.NewRoot())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 3 {
			t.Fatalf("expected 3 roles in root scope, got %d: %+v", len(res), res)
		}
		for _, r := range res {
			if r.Meta["slug"] == "root" {
				t.Fatalf("expected root slug to be excluded from select listing")
			}
		}
	})
}
