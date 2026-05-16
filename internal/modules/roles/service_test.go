package roles

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// fakeRepository is a test double for the repository.
type fakeRepository struct {
	roles        []Role
	roleByName   map[string]*Role
	total        int64
	err          error
	getByNameErr error
	createErr    error
	updateErr    error
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
