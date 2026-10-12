package authenticate_user

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	userdto "github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"github.com/enviniom/nexokit/internal/modules/iam/core/refresh_token"
	"github.com/enviniom/nexokit/internal/platform/identity"
)

type PasswordVerifier interface {
	VerifyPassword(password, hash string) error
}

type Service interface {
	Login(req userdto.LoginRequest) (*userdto.LoginResponse, error)
}

type service struct {
	repo       Repository
	verifier   PasswordVerifier
	manager    core.TokenManager
	refreshTTL time.Duration
}

func NewService(repo Repository, verifier PasswordVerifier, manager core.TokenManager, refreshTTL time.Duration) Service {
	return &service{repo: repo, verifier: verifier, manager: manager, refreshTTL: refreshTTL}
}

func (s *service) Login(req userdto.LoginRequest) (*userdto.LoginResponse, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, core.ErrInvalidCredentials
	}
	if err := s.verifier.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, core.ErrInvalidCredentials
	}

	access, err := s.manager.IssueAccess(user.PublicID, user.Role.Name, user.CompanyID)
	if err != nil {
		return nil, err
	}
	refresh, err := s.manager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateRefreshToken(&refresh_token.RefreshToken{
		PublicID:  publicID,
		UserID:    user.ID,
		TokenHash: s.manager.HashRefreshToken(refresh),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return nil, err
	}

	result := &userdto.LoginResponse{AccessToken: access, RefreshToken: refresh}
	result.User.FromModel(*user)
	return result, nil
}
