package sync_permissions

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"testing"
)

type fakeRepo struct {
	items       map[string]*permission.IAMPermission
	next        uint
	findErr     error
	createErr   error
	assignErr   error
	assignCalls int
}

func (f *fakeRepo) FindBySlug(slug string) (*permission.IAMPermission, bool, error) {
	if f.findErr != nil {
		return nil, false, f.findErr
	}
	if p, ok := f.items[slug]; ok {
		return p, true, nil
	}
	return nil, false, nil
}

func (f *fakeRepo) Create(permission *permission.IAMPermission) error {
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
	repo := &fakeRepo{items: map[string]*permission.IAMPermission{}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"users.create", "users.create"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one permission, got %d", len(repo.items))
	}
}

func TestSyncPermissionsSkipsExisting(t *testing.T) {
	existing := &permission.IAMPermission{Slug: "users.create"}
	existing.ID = 99
	repo := &fakeRepo{items: map[string]*permission.IAMPermission{"users.create": existing}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"users.create"}); err != nil {
		t.Fatal(err)
	}
	if repo.assignCalls != 0 {
		t.Fatalf("expected no assign calls for existing permission, got %d", repo.assignCalls)
	}
}

func TestSyncPermissionsPropagatesFindError(t *testing.T) {
	repo := &fakeRepo{items: map[string]*permission.IAMPermission{}, findErr: errors.New("db down")}
	svc := NewService(repo)
	err := svc.SyncPermissions([]string{"users.create"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSyncPermissionsIgnoresMalformedSlugs(t *testing.T) {
	repo := &fakeRepo{items: map[string]*permission.IAMPermission{}}
	svc := NewService(repo)
	if err := svc.SyncPermissions([]string{"noslash", "valid.create"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one permission, got %d", len(repo.items))
	}
}

func TestSyncPermissionsPropagatesCreateError(t *testing.T) {
	createErr := errors.New("create failed")
	repo := &fakeRepo{items: map[string]*permission.IAMPermission{}, createErr: createErr}
	svc := NewService(repo)
	err := svc.SyncPermissions([]string{"users.create"})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestSyncPermissionsPropagatesAssignError(t *testing.T) {
	assignErr := errors.New("assign failed")
	repo := &fakeRepo{items: map[string]*permission.IAMPermission{}, assignErr: assignErr}
	svc := NewService(repo)
	err := svc.SyncPermissions([]string{"users.create"})
	if !errors.Is(err, assignErr) {
		t.Fatalf("expected assign error, got %v", err)
	}
}
