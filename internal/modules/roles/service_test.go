package roles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/platform/apperror"
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

func (f *fakeRepository) List(page, perPage int) ([]Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.roles, nil
}

func (f *fakeRepository) Count() (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.total, nil
}

func (f *fakeRepository) GetByPublicID(publicID string) (*Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.roles {
		if f.roles[i].PublicID == publicID {
			return &f.roles[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetByName(name string) (*Role, error) {
	if f.getByNameErr != nil {
		return nil, f.getByNameErr
	}
	if f.err != nil {
		return nil, f.err
	}
	if r, ok := f.roleByName[name]; ok {
		return r, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetBySlug(slug string) (*Role, error) {
	if f.getByNameErr != nil {
		return nil, f.getByNameErr
	}
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.roles {
		if f.roles[i].Slug == slug {
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

func (f *fakeRepository) Delete(publicID string) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.roles {
		if f.roles[i].PublicID == publicID {
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

type fakePermissionCatalogRepository struct {
	permissions []permissions.Permission
	err         error
}

func (f fakePermissionCatalogRepository) ListAll() ([]permissions.Permission, error) {
	return f.permissions, f.err
}

type fakeRoleMemberRepository struct {
	publicIDs []string
	err       error
}

func (f fakeRoleMemberRepository) ListPublicIDsByRoleID(roleID uint) ([]string, error) {
	return f.publicIDs, f.err
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

		result, total, err := svc.List(1, 10)
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

		result, total, err := svc.List(1, 10)
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

		_, _, err := svc.List(1, 10)
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

		result, err := svc.GetByPublicID("role1")
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

		_, err := svc.GetByPublicID("missing")
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
		result, err := svc.Create(req)
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
		_, err := svc.Create(req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns repository error when uniqueness check fails", func(t *testing.T) {
		repo := &fakeRepository{getByNameErr: apperror.ErrInternal}
		svc := NewService(repo)

		req := CreateRoleRequest{Name: "editor", Slug: "editor"}
		_, err := svc.Create(req)
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
		_, err := svc.Create(req)
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
		result, err := svc.Update("role1", req)
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
		_, err := svc.Update("role1", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns conflict when updating to an existing name", func(t *testing.T) {
		repo := &fakeRepository{
			roles: []Role{
				{BaseModel: shared.BaseModel{PublicID: "role1"}, Name: "editor", IsSystem: false},
				{BaseModel: shared.BaseModel{PublicID: "role2"}, Name: "admin", IsSystem: false},
			},
			roleByName: map[string]*Role{
				"admin": {BaseModel: shared.BaseModel{PublicID: "role2"}, Name: "admin"},
			},
		}
		svc := NewService(repo)

		req := UpdateRoleRequest{Name: "admin"}
		_, err := svc.Update("role1", req)
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

		req := UpdateRoleRequest{Name: "admin"}
		_, err := svc.Update("role1", req)
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

		req := UpdateRoleRequest{Name: "admin"}
		_, err := svc.Update("role1", req)
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

		err := svc.Delete("role1")
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

		err := svc.Delete("role1")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestService_GetPermissionCatalog(t *testing.T) {
	t.Run("returns grouped catalog with granted flags", func(t *testing.T) {
		repo := &fakeRepository{roles: []Role{{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-admin"}, Name: "admin", Permissions: []permissions.Permission{{Slug: "users.index"}}}}}
		catalog := fakePermissionCatalogRepository{permissions: []permissions.Permission{
			{BaseModel: shared.BaseModel{PublicID: "p1"}, Module: "roles", Slug: "roles.index", Action: permissions.ActionIndex, DisplayOrder: 10},
			{BaseModel: shared.BaseModel{PublicID: "p2"}, Module: "users", Slug: "users.index", Action: permissions.ActionIndex, DisplayOrder: 10},
			{BaseModel: shared.BaseModel{PublicID: "p3"}, Module: "users", Slug: "users.view", Action: permissions.ActionView, DisplayOrder: 20},
		}}
		svc := NewService(repo, WithPermissionCatalog(catalog))

		groups, err := svc.GetPermissionCatalog("role-admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 2 || groups[0].Module != "roles" || groups[1].Module != "users" {
			t.Fatalf("expected roles and users groups, got %+v", groups)
		}
		if !groups[1].Permissions[0].Granted || groups[1].Permissions[1].Granted {
			t.Fatalf("expected granted flag only for users.index, got %+v", groups[1].Permissions)
		}
	})

	t.Run("returns not found for missing role", func(t *testing.T) {
		svc := NewService(&fakeRepository{roles: []Role{}}, WithPermissionCatalog(fakePermissionCatalogRepository{}))
		_, err := svc.GetPermissionCatalog("missing")
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
			role:       Role{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-admin"}, Name: "admin", Permissions: []permissions.Permission{{BaseModel: shared.BaseModel{ID: 1}, Slug: "users.index"}}},
			catalog:    []permissions.Permission{{BaseModel: shared.BaseModel{ID: 2}, Slug: "users.view", Module: "users"}, {BaseModel: shared.BaseModel{ID: 3}, Slug: "roles.index", Module: "roles"}},
			req:        AssignRolePermissionsRequest{Permissions: []string{"users.view", "roles.index"}},
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
			role:       Role{BaseModel: shared.BaseModel{ID: 7, PublicID: "role-root"}, Name: "root", IsSystem: true, Permissions: []permissions.Permission{{BaseModel: shared.BaseModel{ID: 1}, Slug: "users.index", IsSystem: true}}},
			catalog:    []permissions.Permission{{BaseModel: shared.BaseModel{ID: 1}, Slug: "users.index", Module: "users", IsSystem: true}, {BaseModel: shared.BaseModel{ID: 2}, Slug: "users.view", Module: "users"}},
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

			_, err := svc.AssignPermissions(tt.role.PublicID, tt.req, tt.actorPerms)
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
