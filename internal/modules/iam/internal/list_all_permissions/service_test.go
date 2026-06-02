package list_all_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type fakeRepo struct{ items []core.IAMPermission }

func (f fakeRepo) ListAllPermissions() ([]core.IAMPermission, error) { return f.items, nil }

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
