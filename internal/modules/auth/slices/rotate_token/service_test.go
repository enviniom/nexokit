package rotate_token

import (
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct {
	stored  *core.RefreshToken
	created []*core.RefreshToken
	revoked []string
	lastRep *string
}

func (f *fakeRepo) GetByHash(hash string) (*core.RefreshToken, error) {
	if f.stored == nil || f.stored.TokenHash != hash {
		return nil, core.ErrInvalidRefreshToken
	}
	return f.stored, nil
}
func (f *fakeRepo) CreateRefreshToken(refresh *core.RefreshToken) error {
	f.created = append(f.created, refresh)
	return nil
}
func (f *fakeRepo) Revoke(hash string, replacedByHash *string) error {
	f.revoked = append(f.revoked, hash)
	f.lastRep = replacedByHash
	return nil
}

type fakeManager struct{ access string }

func (f fakeManager) IssueAccess(sub, role string, companyID *uint) (string, error) {
	return f.access, nil
}

type fakeRefreshManager struct{}

func (fakeManager) GenerateRefreshToken() (string, error)       { return "new-refresh", nil }
func (fakeManager) HashRefreshToken(refreshToken string) string { return "hash:" + refreshToken }

func TestService_Rotate(t *testing.T) {
	active := core.AuthUser{BaseModel: shared.BaseModel{ID: 7, PublicID: "usr_7"}, Role: core.AuthRole{Name: "admin"}, IsActive: true}

	t.Run("rotates valid token pair", func(t *testing.T) {
		repo := &fakeRepo{stored: &core.RefreshToken{UserID: 7, TokenHash: "hash:old-refresh", ExpiresAt: time.Now().Add(time.Hour), User: active}}
		svc := NewService(repo, fakeManager{access: "new-access"}, time.Hour)

		result, err := svc.Rotate(core.RefreshRequest{RefreshToken: "old-refresh"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "new-access" || result.RefreshToken != "new-refresh" {
			t.Fatalf("unexpected token pair: %#v", result)
		}
		if len(repo.created) != 1 || len(repo.revoked) != 1 || repo.revoked[0] != "hash:old-refresh" {
			t.Fatalf("expected create+revoke flow, got created=%d revoked=%#v", len(repo.created), repo.revoked)
		}
	})

	t.Run("rejects revoked token", func(t *testing.T) {
		now := time.Now()
		repo := &fakeRepo{stored: &core.RefreshToken{TokenHash: "hash:old-refresh", ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &now, User: active}}
		svc := NewService(repo, fakeManager{}, time.Hour)

		_, err := svc.Rotate(core.RefreshRequest{RefreshToken: "old-refresh"})
		if !errors.Is(err, core.ErrInvalidRefreshToken) {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		repo := &fakeRepo{stored: &core.RefreshToken{TokenHash: "hash:old-refresh", ExpiresAt: time.Now().Add(-time.Minute), User: active}}
		svc := NewService(repo, fakeManager{}, time.Hour)

		_, err := svc.Rotate(core.RefreshRequest{RefreshToken: "old-refresh"})
		if !errors.Is(err, core.ErrInvalidRefreshToken) {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})
}
