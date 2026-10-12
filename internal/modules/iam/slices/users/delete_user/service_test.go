package delete_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeServiceRepo struct{ err error }

func (f fakeServiceRepo) Delete(tenant.TenantContext, string) error { return f.err }

func TestDeleteUserServiceSuccess(t *testing.T) {
	svc := NewService(fakeServiceRepo{})

	if err := svc.Delete(tenant.NewRoot(), "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUserServiceNotFoundPropagation(t *testing.T) {
	svc := NewService(fakeServiceRepo{err: core.ErrNotFound})

	err := svc.Delete(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteUserServiceUnexpectedRepoError(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	svc := NewService(fakeServiceRepo{err: repoErr})

	err := svc.Delete(tenant.NewRoot(), "user-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
