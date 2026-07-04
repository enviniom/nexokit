package view_permission

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeServiceRepo struct {
	item *core.IAMPermission
	err  error
}

func (f fakeServiceRepo) GetPermissionByPublicID(string) (*core.IAMPermission, error) {
	return f.item, f.err
}

func TestGetByPublicIDSuccess(t *testing.T) {
	svc := NewService(fakeServiceRepo{item: &core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-1"}, Slug: "permissions.manage", Name: "Manage", Module: "permissions", Action: "manage"}})

	item, err := svc.GetByPublicID("perm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID != "perm-1" {
		t.Fatalf("expected public id perm-1, got %s", item.PublicID)
	}
}

func TestGetByPublicIDPropagatesNotFound(t *testing.T) {
	svc := NewService(fakeServiceRepo{err: core.ErrNotFound})
	_, err := svc.GetByPublicID("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetByPublicIDPropagatesUnexpectedError(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	svc := NewService(fakeServiceRepo{err: repoErr})
	_, err := svc.GetByPublicID("perm-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
