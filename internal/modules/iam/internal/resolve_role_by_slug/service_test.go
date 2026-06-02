package resolve_role_by_slug

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type fakeRepo struct {
	role *core.IAMRole
	err  error
}

func (f fakeRepo) GetRoleBySlug(string) (*core.IAMRole, error) { return f.role, f.err }

func TestResolveRoleBySlugMapsNotFound(t *testing.T) {
	svc := NewService(fakeRepo{err: gorm.ErrRecordNotFound})
	_, err := svc.ResolveRoleBySlug("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
