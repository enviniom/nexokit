package rotate_token

import (
	"errors"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"gorm.io/gorm"
)

type Service interface {
	Rotate(req core.RefreshRequest) (*core.TokenPairResponse, error)
}

type service struct {
	repo         Repository
	tokenManager core.TokenManager
	refreshTTL   time.Duration
}

func NewService(repo Repository, tokenManager core.TokenManager, refreshTTL time.Duration) Service {
	return &service{repo: repo, tokenManager: tokenManager, refreshTTL: refreshTTL}
}

func (s *service) Rotate(req core.RefreshRequest) (*core.TokenPairResponse, error) {
	hash := s.tokenManager.HashRefreshToken(req.RefreshToken)
	stored, err := s.repo.GetByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized
		}
		return nil, err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) || !stored.User.IsActive {
		return nil, apperror.ErrUnauthorized
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
	if err := s.repo.CreateRefreshToken(&core.RefreshToken{PublicID: publicID, UserID: stored.UserID, TokenHash: s.tokenManager.HashRefreshToken(refresh), ExpiresAt: time.Now().Add(s.refreshTTL)}); err != nil {
		return nil, err
	}
	replacement := s.tokenManager.HashRefreshToken(refresh)
	if err := s.repo.Revoke(hash, &replacement); err != nil {
		return nil, err
	}

	return &core.TokenPairResponse{AccessToken: access, RefreshToken: refresh}, nil
}
