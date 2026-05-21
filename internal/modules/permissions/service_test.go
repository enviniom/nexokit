package permissions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

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
	userSlugs   map[string][]string
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

func (f *fakeRepository) ListSlugsByUserPublicID(publicID string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]string(nil), f.userSlugs[publicID]...), nil
}

type resolverCache struct {
	values map[string][]byte
	sets   map[string][]byte
	ttls   map[string]time.Duration
}

func (c *resolverCache) Get(ctx context.Context, key string) ([]byte, error) {
	if c.values == nil {
		return nil, nil
	}
	return c.values[key], nil
}

func (c *resolverCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if c.sets == nil {
		c.sets = make(map[string][]byte)
	}
	if c.ttls == nil {
		c.ttls = make(map[string]time.Duration)
	}
	c.sets[key] = value
	c.ttls[key] = ttl
	return nil
}

func (c *resolverCache) Delete(ctx context.Context, key string) error { return nil }
func (c *resolverCache) Exists(ctx context.Context, key string) (bool, error) {
	if c.values == nil {
		return false, nil
	}
	_, ok := c.values[key]
	return ok, nil
}
func (c *resolverCache) Close() error { return nil }

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

func TestService_ResolvePermissions(t *testing.T) {
	t.Run("returns cached permissions without repository lookup", func(t *testing.T) {
		cached, _ := json.Marshal([]string{"users.index", "roles.index"})
		repo := &fakeRepository{err: errors.New("repository should not be called")}
		svc := NewService(repo, WithCache(&resolverCache{values: map[string][]byte{"rbac:permissions:user-one": cached}}))

		got, err := svc.Resolve("user-one")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 || got[0] != "users.index" || got[1] != "roles.index" {
			t.Fatalf("expected cached slugs, got %v", got)
		}
	})

	t.Run("loads permissions from repository and stores five minute cache entry", func(t *testing.T) {
		cache := &resolverCache{}
		repo := &fakeRepository{userSlugs: map[string][]string{"user-one": {"users.view", "auth.view"}}}
		svc := NewService(repo, WithCache(cache))

		got, err := svc.Resolve("user-one")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 || got[0] != "users.view" || got[1] != "auth.view" {
			t.Fatalf("expected repository slugs, got %v", got)
		}
		if cache.ttls["rbac:permissions:user-one"] != 5*time.Minute {
			t.Fatalf("expected 5 minute ttl, got %v", cache.ttls["rbac:permissions:user-one"])
		}
		var stored []string
		if err := json.Unmarshal(cache.sets["rbac:permissions:user-one"], &stored); err != nil {
			t.Fatalf("cached value is not a string slice: %v", err)
		}
		if len(stored) != 2 || stored[0] != "users.view" || stored[1] != "auth.view" {
			t.Fatalf("expected cached repository slugs, got %v", stored)
		}
	})
}
