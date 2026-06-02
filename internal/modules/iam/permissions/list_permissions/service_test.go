package list_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct{ items []core.IAMPermission }

func (f fakeRepo) ListAllPermissions() ([]core.IAMPermission, error) { return f.items, nil }

func TestListGrouped(t *testing.T) {
	svc := NewService(fakeRepo{items: []core.IAMPermission{
		{BaseModel: shared.BaseModel{PublicID: "2"}, Module: "users", Slug: "users.view", DisplayOrder: 20},
		{BaseModel: shared.BaseModel{PublicID: "1"}, Module: "roles", Slug: "roles.list", DisplayOrder: 10},
	}})

	g, err := svc.ListGrouped()
	if err != nil {
		t.Fatal(err)
	}
	if len(g) != 2 || g[0].Module != "roles" {
		t.Fatalf("unexpected grouping: %+v", g)
	}
}
