package auth

import (
	"errors"
	"time"

	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"gorm.io/gorm"
)

// PasswordVerifier verifies user passwords.
type PasswordVerifier interface {
	VerifyPassword(password, hash string) error
}

// AccessIssuer issues access tokens.
type AccessIssuer interface {
	IssueAccess(sub, role string, companyID *uint) (string, error)
}

// RefreshTokenManager generates and hashes opaque refresh tokens.
type RefreshTokenManager interface {
	GenerateRefreshToken() (string, error)
	HashRefreshToken(refreshToken string) string
}

// Service defines the auth business logic contract.
type Service interface {
	Login(req LoginRequest) (*LoginResponse, error)
	Refresh(req RefreshRequest) (*TokenPairResponse, error)
	Logout(req RefreshRequest) error
}

type service struct {
	usersRepo     users.Repository
	verifier      PasswordVerifier
	issuer        AccessIssuer
	refreshTokens RefreshTokenManager
	refreshes     RefreshRepository
	refreshTTL    time.Duration
}

// NewService creates a new auth service.
func NewService(usersRepo users.Repository, verifier PasswordVerifier, issuer AccessIssuer, refreshTokens RefreshTokenManager, refreshes RefreshRepository, refreshTTL time.Duration) Service {
	return &service{usersRepo: usersRepo, verifier: verifier, issuer: issuer, refreshTokens: refreshTokens, refreshes: refreshes, refreshTTL: refreshTTL}
}

// Login verifies credentials and issues an access/refresh token pair.
func (s *service) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.usersRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, apperror.ErrUnauthorized
	}
	if err := s.verifier.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, apperror.ErrUnauthorized
	}

	access, refresh, err := s.issueAndStoreTokenPair(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{AccessToken: access, RefreshToken: refresh, User: toUserResponse(user)}, nil
}

// Refresh validates a refresh token, revokes it, and returns a rotated token pair.
func (s *service) Refresh(req RefreshRequest) (*TokenPairResponse, error) {
	hash := s.refreshTokens.HashRefreshToken(req.RefreshToken)
	stored, err := s.refreshes.GetByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized
		}
		return nil, err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) || !stored.User.IsActive {
		return nil, apperror.ErrUnauthorized
	}

	access, refresh, err := s.issueAndStoreTokenPair(&stored.User)
	if err != nil {
		return nil, err
	}
	replacement := s.refreshTokens.HashRefreshToken(refresh)
	if err := s.refreshes.Revoke(hash, &replacement); err != nil {
		return nil, err
	}

	return &TokenPairResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// Logout revokes the provided refresh token when it exists and is usable.
func (s *service) Logout(req RefreshRequest) error {
	hash := s.refreshTokens.HashRefreshToken(req.RefreshToken)
	stored, err := s.refreshes.GetByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrUnauthorized
		}
		return err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return apperror.ErrUnauthorized
	}
	return s.refreshes.Revoke(hash, nil)
}

func (s *service) issueAndStoreTokenPair(user *users.User) (string, string, error) {
	access, err := s.issuer.IssueAccess(user.PublicID, user.Role.Name, user.CompanyID)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.refreshTokens.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	publicID, err := identity.Generate()
	if err != nil {
		return "", "", err
	}
	if err := s.refreshes.Create(&RefreshToken{
		PublicID:  publicID,
		UserID:    user.ID,
		TokenHash: s.refreshTokens.HashRefreshToken(refresh),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func toUserResponse(user *users.User) users.UserResponse {
	return users.UserResponse{
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
