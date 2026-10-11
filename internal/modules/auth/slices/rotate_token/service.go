package rotate_token

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/modules/auth/core/refresh_token"
	"github.com/enviniom/nexokit/internal/platform/identity"
)

type Service interface {
	Rotate(req refresh_token.RefreshRequest) (*refresh_token.TokenPairResponse, error)
}

type service struct {
	repo         Repository
	tokenManager core.TokenManager
	refreshTTL   time.Duration
}

func NewService(repo Repository, tokenManager core.TokenManager, refreshTTL time.Duration) Service {
	return &service{repo: repo, tokenManager: tokenManager, refreshTTL: refreshTTL}
}

func (s *service) Rotate(req refresh_token.RefreshRequest) (*refresh_token.TokenPairResponse, error) {
	hash := s.tokenManager.HashRefreshToken(req.RefreshToken)
	stored, err := s.repo.GetByHash(hash)
	if err != nil {
		return nil, err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) || !stored.User.IsActive {
		return nil, core.ErrInvalidRefreshToken
	}

	access, err := s.tokenManager.IssueAccess(stored.User.PublicID, stored.User.Role.Name, stored.User.CompanyID)
	if err != nil {
		return nil, err
	}
	refresh, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateRefreshToken(&refresh_token.RefreshToken{PublicID: publicID, UserID: stored.UserID, TokenHash: s.tokenManager.HashRefreshToken(refresh), ExpiresAt: time.Now().Add(s.refreshTTL)}); err != nil {
		return nil, err
	}
	replacement := s.tokenManager.HashRefreshToken(refresh)
	if err := s.repo.Revoke(hash, &replacement); err != nil {
		return nil, err
	}

	return &refresh_token.TokenPairResponse{AccessToken: access, RefreshToken: refresh}, nil
}
