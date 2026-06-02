package resolve_role_by_slug

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type fakeRepo struct {
	role *core.IAMRole
	err  error
}

func (f fakeRepo) GetRoleBySlug(string) (*core.IAMRole, error) { return f.role, f.err }

func TestResolveRoleBySlugPropagatesNotFound(t *testing.T) {
	svc := NewService(fakeRepo{err: core.ErrNotFound})
	_, err := svc.ResolveRoleBySlug("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveRoleBySlugSuccess(t *testing.T) {
	role := &core.IAMRole{Name: "Admin", Slug: "admin"}
	svc := NewService(fakeRepo{role: role})
	result, err := svc.ResolveRoleBySlug("admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "admin" {
		t.Fatalf("expected slug admin, got %s", result.Slug)
	}
}
