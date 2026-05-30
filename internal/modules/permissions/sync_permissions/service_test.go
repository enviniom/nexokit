package sync_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

type fakeRepo struct {
	bySlug map[string]*core.Permission
	lastID uint
}

func (f *fakeRepo) GetBySlug(slug string) (*core.Permission, error) {
	if p, ok := f.bySlug[slug]; ok { return p, nil }
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeRepo) Create(permission *core.Permission) error {
	f.lastID++
	permission.ID = f.lastID
	f.bySlug[permission.Slug] = permission
	return nil
}
func (f *fakeRepo) AutoAssignToAdmins(uint) error { return nil }

func TestSyncPermissionsIdempotent(t *testing.T) {
	repo := &fakeRepo{bySlug: map[string]*core.Permission{"users.list": {Slug: "users.list"}}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"users.list", "roles.list"}); err != nil { t.Fatal(err) }
	if len(repo.bySlug) != 2 { t.Fatalf("expected 2 permissions, got %d", len(repo.bySlug)) }
}
