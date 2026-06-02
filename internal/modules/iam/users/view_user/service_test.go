package view_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeServiceRepo struct {
	item *core.UserResponse
	err  error
}

func (f fakeServiceRepo) GetByPublicID(tenant.TenantContext, string) (*core.UserResponse, error) {
	return f.item, f.err
}

func TestViewUserServiceSuccess(t *testing.T) {
	svc := NewService(fakeServiceRepo{item: &core.UserResponse{PublicID: "user-1", Name: "Alice"}})

	item, err := svc.GetByPublicID(tenant.NewRoot(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID != "user-1" {
		t.Fatalf("expected user-1, got %s", item.PublicID)
	}
	if item.Name != "Alice" {
		t.Fatalf("expected name Alice, got %s", item.Name)
	}
}

func TestViewUserServiceNotFoundPropagation(t *testing.T) {
	svc := NewService(fakeServiceRepo{err: core.ErrNotFound})

	_, err := svc.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestViewUserServiceUnexpectedRepoError(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	svc := NewService(fakeServiceRepo{err: repoErr})

	_, err := svc.GetByPublicID(tenant.NewRoot(), "user-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
