package revoke_token

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/core/refresh_token"
)

type RefreshTokenManager interface {
	HashRefreshToken(refreshToken string) string
}

type Service interface {
	Revoke(req refresh_token.RefreshRequest) error
}

type service struct {
	repo          Repository
	refreshTokens RefreshTokenManager
}

func NewService(repo Repository, refreshTokens RefreshTokenManager) Service {
	return &service{repo: repo, refreshTokens: refreshTokens}
}

func (s *service) Revoke(req refresh_token.RefreshRequest) error {
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
