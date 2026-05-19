package permissions

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeRepository struct {
	permissions []Permission
	err         error
	createErr   error
	updateErr   error
}

func (f *fakeRepository) List(page, perPage int) ([]Permission, error) {
	return f.permissions, f.err
}

func (f *fakeRepository) ListAll() ([]Permission, error) {
	return f.permissions, f.err
}

func (f *fakeRepository) Count() (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return int64(len(f.permissions)), nil
}

func (f *fakeRepository) GetByPublicID(publicID string) (*Permission, error) {
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.permissions {
		if f.permissions[i].PublicID == publicID {
			return &f.permissions[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetBySlug(slug string) (*Permission, error) {
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.permissions {
		if f.permissions[i].Slug == slug {
			return &f.permissions[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) Create(permission *Permission) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.permissions = append(f.permissions, *permission)
	return nil
}

func (f *fakeRepository) Update(permission *Permission) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	for i := range f.permissions {
		if f.permissions[i].PublicID == permission.PublicID {
			f.permissions[i] = *permission
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeRepository) Delete(publicID string) error {
	for i := range f.permissions {
		if f.permissions[i].PublicID == publicID {
			f.permissions = append(f.permissions[:i], f.permissions[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func TestPermissionFieldsAndSlugUniqueness(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&Permission{}); err != nil {
		t.Fatalf("failed to migrate permissions: %v", err)
	}

	repo := NewRepository(db)
	permission := &Permission{
		BaseModel:    shared.BaseModel{PublicID: "perm_users_create_000001"},
		Slug:         "users.create",
		Name:         "Create users",
		Module:       "users",
		Action:       ActionCreate,
		Description:  "Allows creating users",
		IsSystem:     true,
		DisplayOrder: 30,
	}
	if err := repo.Create(permission); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	stored, err := repo.GetBySlug("users.create")
	if err != nil {
		t.Fatalf("expected permission by slug: %v", err)
	}
	if stored.Module != "users" || stored.Action != ActionCreate || stored.DisplayOrder != 30 || !stored.IsSystem {
		t.Fatalf("stored permission fields mismatch: %+v", stored)
	}

	duplicate := *permission
	duplicate.ID = 0
	duplicate.PublicID = "perm_users_create_000002"
	if err := repo.Create(&duplicate); err == nil {
		t.Fatal("expected unique constraint error for duplicate slug")
	}
}

func TestValidatePermissionParts(t *testing.T) {
	tests := []struct {
		name    string
		module  string
		action  string
		slug    string
		wantErr bool
	}{
		{name: "explicit create action", module: "users", action: ActionCreate, slug: "users.create"},
		{name: "business change role action", module: "users", action: ActionChangeRole, slug: "users.change_role"},
		{name: "business manage action", module: "permissions", action: ActionManage, slug: "permissions.manage"},
		{name: "reject ambiguous read action", module: "users", action: "read", slug: "users.read", wantErr: true},
		{name: "reject slug mismatch", module: "users", action: ActionCreate, slug: "users.update", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePermissionParts(tt.module, tt.action, tt.slug)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_ListGroupedSortsByModuleAndDisplayOrder(t *testing.T) {
	repo := &fakeRepository{permissions: []Permission{
		{BaseModel: shared.BaseModel{PublicID: "p3"}, Module: "users", Slug: "users.delete", Action: ActionDelete, DisplayOrder: 50},
		{BaseModel: shared.BaseModel{PublicID: "p1"}, Module: "roles", Slug: "roles.index", Action: ActionIndex, DisplayOrder: 10},
		{BaseModel: shared.BaseModel{PublicID: "p2"}, Module: "users", Slug: "users.index", Action: ActionIndex, DisplayOrder: 10},
	}}
	svc := NewService(repo)

	groups, err := svc.ListGrouped()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Module != "roles" || groups[1].Module != "users" {
		t.Fatalf("expected groups sorted by module, got %+v", groups)
	}
	if got := groups[1].Permissions[0].Slug; got != "users.index" {
		t.Fatalf("expected users.index first by display_order, got %s", got)
	}
}

func TestService_SystemCRUDProtection(t *testing.T) {
	tests := []struct {
		name string
		run  func(Service) error
	}{
		{
			name: "update system permission forbidden",
			run: func(svc Service) error {
				_, err := svc.Update("system", UpdatePermissionRequest{Slug: "users.index", Name: "List users", Module: "users", Action: ActionIndex})
				return err
			},
		},
		{
			name: "delete system permission forbidden",
			run:  func(svc Service) error { return svc.Delete("system") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{permissions: []Permission{{BaseModel: shared.BaseModel{PublicID: "system"}, Slug: "users.index", Module: "users", Action: ActionIndex, IsSystem: true}}}
			err := tt.run(NewService(repo))
			if !errors.Is(err, apperror.ErrForbidden) {
				t.Fatalf("expected forbidden, got %v", err)
			}
		})
	}
}
