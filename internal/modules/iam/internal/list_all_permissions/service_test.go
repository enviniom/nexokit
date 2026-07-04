package list_all_permissions

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type fakeRepo struct {
	items []core.IAMPermission
	err   error
}

func (f fakeRepo) ListAllPermissions() ([]core.IAMPermission, error) { return f.items, f.err }

func TestListAllPermissions(t *testing.T) {
	want := []core.IAMPermission{{Slug: "users.view"}}
	svc := NewService(fakeRepo{items: want})
	got, err := svc.ListAllPermissions()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Slug != "users.view" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestListAllPermissionsPropagatesError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewService(fakeRepo{err: repoErr})
	_, err := svc.ListAllPermissions()
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
