package revoke_token

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
)

type RefreshTokenManager interface {
	HashRefreshToken(refreshToken string) string
}

type Service interface {
	Revoke(req core.RefreshRequest) error
}

type service struct {
	repo          Repository
	refreshTokens RefreshTokenManager
}

func NewService(repo Repository, refreshTokens RefreshTokenManager) Service {
	return &service{repo: repo, refreshTokens: refreshTokens}
}

func (s *service) Revoke(req core.RefreshRequest) error {
	hash := s.refreshTokens.HashRefreshToken(req.RefreshToken)
	stored, err := s.repo.GetByHash(hash)
	if err != nil {
		return err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return core.ErrInvalidRefreshToken
	}
	return s.repo.Revoke(hash)
}
