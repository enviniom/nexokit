package toggle_user_status

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeRepo struct {
	resp *core.UserResponse
	err  error
}

func (f fakeRepo) ToggleStatus(tenant.TenantContext, string, bool) (*core.UserResponse, error) {
	return f.resp, f.err
}

func TestToggleServiceSuccess(t *testing.T) {
	expected := &core.UserResponse{PublicID: "user-1", Name: "Alice", IsActive: false}
	svc := NewService(fakeRepo{resp: expected})

	resp, err := svc.Toggle(tenant.NewRoot(), "user-1", core.UpdateStatusRequest{IsActive: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PublicID != "user-1" {
		t.Fatalf("expected user-1, got %s", resp.PublicID)
	}
	if resp.IsActive {
		t.Fatalf("expected IsActive false, got true")
	}
}

func TestToggleServiceNotFound(t *testing.T) {
	svc := NewService(fakeRepo{err: core.ErrNotFound})

	_, err := svc.Toggle(tenant.NewRoot(), "missing", core.UpdateStatusRequest{IsActive: true})
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestToggleServiceUpdateError(t *testing.T) {
	repoErr := errors.New("db write failure")
	svc := NewService(fakeRepo{err: repoErr})

	_, err := svc.Toggle(tenant.NewRoot(), "user-1", core.UpdateStatusRequest{IsActive: false})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
