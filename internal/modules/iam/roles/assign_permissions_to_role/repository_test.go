package assign_permissions_to_role

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeCache struct{ deleted []string }

func (f *fakeCache) Get(context.Context, string) ([]byte, error)              { return nil, nil }
func (f *fakeCache) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (f *fakeCache) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}
func (f *fakeCache) Exists(context.Context, string) (bool, error) { return false, nil }
func (f *fakeCache) Close() error                                 { return nil }

func TestRepositoryAssignFlow(t *testing.T) {
	db := mustOpenAssignDB(t)
	company, role, permissions := seedAssignData(t, db)

	repo := NewRepository(db)
	tc := tenant.NewScoped(company.ID, company.Slug)

	t.Run("maps missing role to domain error", func(t *testing.T) {
		_, err := repo.GetByPublicID(tc, "missing")
		if !errors.Is(err, core.ErrNotFound) {
			t.Fatalf("expected core.ErrNotFound, got %v", err)
		}
	})

	t.Run("resolves and replaces permissions", func(t *testing.T) {
		catalog, err := repo.ListAllPermissions()
		if err != nil {
			t.Fatalf("catalog: %v", err)
		}
		normalized, selected, ids, err := repo.ResolvePermissionSelection(catalog, []string{permissions[0].Slug, permissions[1].Slug})
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if len(normalized) != 2 {
			t.Fatalf("expected 2 normalized slugs, got %d", len(normalized))
		}
		if !selected[permissions[0].Slug] || !selected[permissions[1].Slug] {
			t.Fatalf("expected selected map to contain both permissions")
		}
		if err := repo.ReplacePermissions(role.ID, ids); err != nil {
			t.Fatalf("replace: %v", err)
		}

		var links []core.IAMRolePermission
		if err := db.Where("role_id = ?", role.ID).Order("permission_id asc").Find(&links).Error; err != nil {
			t.Fatalf("load links: %v", err)
		}
		if len(links) != 2 {
			t.Fatalf("expected 2 links, got %d", len(links))
		}
	})

	t.Run("invalidates member cache keys", func(t *testing.T) {
		cache := &fakeCache{}
		if err := repo.InvalidateRoleMemberPermissionCache(role.ID, cache); err != nil {
			t.Fatalf("invalidate cache: %v", err)
		}
		if len(cache.deleted) != 2 {
			t.Fatalf("expected 2 cache keys deleted, got %d", len(cache.deleted))
		}
	})
}

func mustOpenAssignDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMCompany{}, &core.IAMRole{}, &core.IAMPermission{}, &core.IAMRolePermission{}, &core.IAMUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedAssignData(t *testing.T, db *gorm.DB) (core.IAMCompany, core.IAMRole, []core.IAMPermission) {
	t.Helper()
	company := core.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: "comp-1"}, Name: "Acme", Slug: "acme"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	permissions := []core.IAMPermission{
		{BaseModel: shared.BaseModel{PublicID: "perm-1"}, Slug: "roles.read", Module: "roles", Action: "read", Name: "Read", DisplayOrder: 1},
		{BaseModel: shared.BaseModel{PublicID: "perm-2"}, Slug: "roles.write", Module: "roles", Action: "write", Name: "Write", DisplayOrder: 2},
	}
	for i := range permissions {
		if err := db.Create(&permissions[i]).Error; err != nil {
			t.Fatalf("permission %s: %v", permissions[i].Slug, err)
		}
	}
	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Manager", Slug: "manager", CompanyID: &company.ID}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("role: %v", err)
	}
	users := []core.IAMUser{
		{BaseModel: shared.BaseModel{PublicID: "user-1"}, Name: "A", Email: "a@test.dev", PasswordHash: "hash", RoleID: role.ID, IsActive: true},
		{BaseModel: shared.BaseModel{PublicID: "user-2"}, Name: "B", Email: "b@test.dev", PasswordHash: "hash", RoleID: role.ID, IsActive: true},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("user %s: %v", users[i].PublicID, err)
		}
	}
	return company, role, permissions
}
