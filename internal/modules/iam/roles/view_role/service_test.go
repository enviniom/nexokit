package view_role

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeServiceRepo struct {
	item *core.RoleResponse
	err  error
}

func (f fakeServiceRepo) GetByPublicID(tenant.TenantContext, string) (*core.RoleResponse, error) {
	return f.item, f.err
}

func TestViewRoleServiceSuccess(t *testing.T) {
	svc := NewService(fakeServiceRepo{item: &core.RoleResponse{PublicID: "role-1", Name: "Admin"}})

	item, err := svc.View(tenant.NewRoot(), "role-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID != "role-1" {
		t.Fatalf("expected role-1, got %s", item.PublicID)
	}
}

func TestViewRoleServiceNotFoundPropagation(t *testing.T) {
	svc := NewService(fakeServiceRepo{err: core.ErrNotFound})

	_, err := svc.View(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestViewRoleServiceUnexpectedRepoError(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	svc := NewService(fakeServiceRepo{err: repoErr})

	_, err := svc.View(tenant.NewRoot(), "role-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
