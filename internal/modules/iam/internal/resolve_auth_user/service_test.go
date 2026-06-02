package resolve_auth_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type fakeRepo struct {
	user  *core.IAMUser
	slugs []string
	err   error
}

func (f fakeRepo) GetAuthUser(string) (*core.IAMUser, error)                  { return f.user, f.err }
func (f fakeRepo) ListPermissionSlugsByUserPublicID(string) ([]string, error) { return f.slugs, nil }

func TestResolveAuthUserMapsNotFound(t *testing.T) {
	svc := NewService(fakeRepo{err: gorm.ErrRecordNotFound})
	_, err := svc.ResolveAuthUser("u1")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
