package sync_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type fakeRepo struct {
	items map[string]*core.IAMPermission
	next  uint
}

func (f *fakeRepo) GetBySlug(slug string) (*core.IAMPermission, error) {
	if p, ok := f.items[slug]; ok {
		return p, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeRepo) Create(permission *core.IAMPermission) error {
	f.next++
	permission.ID = f.next
	f.items[permission.Slug] = permission
	return nil
}
func (f *fakeRepo) AutoAssignToAdmins(uint) error { return nil }

func TestSyncPermissionsIsIdempotent(t *testing.T) {
	repo := &fakeRepo{items: map[string]*core.IAMPermission{}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"users.create", "users.create"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one permission, got %d", len(repo.items))
	}
}
