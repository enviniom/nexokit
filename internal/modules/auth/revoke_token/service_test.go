package revoke_token

import (
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"gorm.io/gorm"
)

type fakeRepo struct {
	stored  *core.RefreshToken
	err     error
	revoked []string
}

func (f *fakeRepo) GetByHash(hash string) (*core.RefreshToken, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.stored == nil || f.stored.TokenHash != hash {
		return nil, gorm.ErrRecordNotFound
	}
	return f.stored, nil
}

func (f *fakeRepo) Revoke(hash string) error {
	f.revoked = append(f.revoked, hash)
	return nil
}

type fakeRefreshManager struct{}
func (fakeRefreshManager) HashRefreshToken(refreshToken string) string { return "hash:" + refreshToken }

func TestService_Revoke(t *testing.T) {
	t.Run("revokes valid token", func(t *testing.T) {
		repo := &fakeRepo{stored: &core.RefreshToken{TokenHash: "hash:refresh", ExpiresAt: time.Now().Add(time.Hour)}}
		svc := NewService(repo, fakeRefreshManager{})

		if err := svc.Revoke(core.RefreshRequest{RefreshToken: "refresh"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.revoked) != 1 || repo.revoked[0] != "hash:refresh" {
			t.Fatalf("expected token to be revoked, got %#v", repo.revoked)
		}
	})

	t.Run("rejects already revoked token", func(t *testing.T) {
		now := time.Now()
		repo := &fakeRepo{stored: &core.RefreshToken{TokenHash: "hash:refresh", ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &now}}
		svc := NewService(repo, fakeRefreshManager{})

		err := svc.Revoke(core.RefreshRequest{RefreshToken: "refresh"})
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected unauthorized, got %v", err)
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		svc := NewService(&fakeRepo{}, fakeRefreshManager{})

		err := svc.Revoke(core.RefreshRequest{RefreshToken: "missing"})
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected unauthorized, got %v", err)
		}
	})
}
