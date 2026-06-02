package sync_permissions

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type fakeRepo struct {
	items      map[string]*core.IAMPermission
	next       uint
	findErr    error
	createErr  error
	assignErr  error
	assignCalls int
}

func (f *fakeRepo) FindBySlug(slug string) (*core.IAMPermission, bool, error) {
	if f.findErr != nil {
		return nil, false, f.findErr
	}
	if p, ok := f.items[slug]; ok {
		return p, true, nil
	}
	return nil, false, nil
}

func (f *fakeRepo) Create(permission *core.IAMPermission) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.next++
	permission.ID = f.next
	f.items[permission.Slug] = permission
	return nil
}

func (f *fakeRepo) AutoAssignToAdmins(uint) error {
	f.assignCalls++
	return f.assignErr
}

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

func TestSyncPermissionsSkipsExisting(t *testing.T) {
	existing := &core.IAMPermission{Slug: "users.create"}
	existing.ID = 99
	repo := &fakeRepo{items: map[string]*core.IAMPermission{"users.create": existing}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"users.create"}); err != nil {
		t.Fatal(err)
	}
	if repo.assignCalls != 0 {
		t.Fatalf("expected no assign calls for existing permission, got %d", repo.assignCalls)
	}
}

func TestSyncPermissionsPropagatesFindError(t *testing.T) {
	repo := &fakeRepo{items: map[string]*core.IAMPermission{}, findErr: errors.New("db down")}
	svc := NewService(repo)
	err := svc.SyncPermissions([]string{"users.create"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSyncPermissionsIgnoresMalformedSlugs(t *testing.T) {
	repo := &fakeRepo{items: map[string]*core.IAMPermission{}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"noslash", "valid.create"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one permission, got %d", len(repo.items))
	}
}
