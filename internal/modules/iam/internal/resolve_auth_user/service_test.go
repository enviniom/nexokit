package resolve_auth_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type fakeRepo struct {
	user  *core.IAMUser
	slugs []string
	err   error
}

func (f fakeRepo) GetAuthUser(string) (*core.IAMUser, error)                  { return f.user, f.err }
func (f fakeRepo) ListPermissionSlugsByUserPublicID(string) ([]string, error) { return f.slugs, nil }

func TestResolveAuthUserPropagatesNotFound(t *testing.T) {
	svc := NewService(fakeRepo{err: core.ErrNotFound})
	_, err := svc.ResolveAuthUser("u1")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveAuthUserSuccess(t *testing.T) {
	user := &core.IAMUser{
		Email: "test@example.com",
		Name:  "Test",
		Role: core.IAMRole{
			Name: "Admin",
			Slug: "admin",
			Permissions: []core.IAMPermission{
				{Slug: "users.create"},
			},
		},
		IsActive: true,
	}
	svc := NewService(fakeRepo{user: user})
	result, err := svc.ResolveAuthUser("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", result.Email)
	}
	if len(result.Permissions) != 1 || result.Permissions[0] != "users.create" {
		t.Fatalf("expected [users.create], got %v", result.Permissions)
	}
}
