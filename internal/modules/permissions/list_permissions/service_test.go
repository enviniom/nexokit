package list_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct{ items []core.Permission }

func (f fakeRepo) ListAll() ([]core.Permission, error) { return f.items, nil }

func TestServiceListGrouped(t *testing.T) {
	svc := NewService(fakeRepo{items: []core.Permission{
		{BaseModel: shared.BaseModel{PublicID: "2"}, Module: "users", Slug: "users.view", DisplayOrder: 20},
		{BaseModel: shared.BaseModel{PublicID: "1"}, Module: "roles", Slug: "roles.list", DisplayOrder: 10},
	}})

	groups, err := svc.ListGrouped()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 || groups[0].Module != "roles" {
		t.Fatalf("unexpected grouping: %+v", groups)
	}
}
