package authenticate_user

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/identity"
)

type PasswordVerifier interface {
	VerifyPassword(password, hash string) error
}

type Service interface {
	Login(req core.LoginRequest) (*core.LoginResponse, error)
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

func (s *service) Login(req core.LoginRequest) (*core.LoginResponse, error) {
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
	if err := s.repo.CreateRefreshToken(&core.RefreshToken{
		PublicID:  publicID,
		UserID:    user.ID,
		TokenHash: s.manager.HashRefreshToken(refresh),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return nil, err
	}

	return &core.LoginResponse{AccessToken: access, RefreshToken: refresh, User: toUserResponse(user)}, nil
}

func toUserResponse(user *core.AuthUser) core.AuthUserResponse {
	return core.AuthUserResponse{
		PublicID:  user.PublicID,
		Name:      user.Name,
		Email:     user.Email,
		IsActive:  user.IsActive,
		RoleID:    user.RoleID,
		RoleName:  user.Role.Name,
		CompanyID: user.CompanyID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		CreatedBy: user.CreatedBy,
		UpdatedBy: user.UpdatedBy,
	}
}
