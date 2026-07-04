package authenticate_user

import (
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepository struct {
	user   *core.AuthUser
	err    error
	stored []*core.RefreshToken
}

func (f *fakeRepository) GetByEmail(email string) (*core.AuthUser, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.user == nil {
		return nil, core.ErrInvalidCredentials
	}
	return f.user, nil
}
func (f *fakeRepository) CreateRefreshToken(refresh *core.RefreshToken) error {
	f.stored = append(f.stored, refresh)
	return nil
}

type fakePasswordVerifier struct{ err error }

func (f fakePasswordVerifier) VerifyPassword(password, hash string) error { return f.err }

type fakeManager struct {
	access string
	err    error
	token  string
}

func (f fakeManager) IssueAccess(sub, role string, companyID *uint) (string, error) {
	return f.access, f.err
}

func (f fakeManager) GenerateRefreshToken() (string, error) { return f.token, nil }
func (f fakeManager) HashRefreshToken(token string) string  { return "hash:" + token }

func activeUser() *core.AuthUser {
	return &core.AuthUser{
		BaseModel:    shared.BaseModel{ID: 7, PublicID: "user-public-id"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		RoleID:       3,
		Role:         core.AuthRole{Name: "admin"},
		IsActive:     true,
	}
}

func TestService_Login(t *testing.T) {
	t.Run("issues access and opaque refresh for active user", func(t *testing.T) {
		repo := &fakeRepository{user: activeUser()}
		svc := NewService(repo, fakePasswordVerifier{}, fakeManager{access: "access-token", token: "refresh-token"}, 7*24*time.Hour)

		result, err := svc.Login(core.LoginRequest{Email: "alice@example.com", Password: "Secret1!"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "access-token" || result.RefreshToken != "refresh-token" {
			t.Fatalf("expected issued tokens, got %#v", result)
		}
		if result.User.Email != "alice@example.com" || result.User.RoleName != "admin" {
			t.Fatalf("expected sanitized user DTO, got %#v", result.User)
		}
		if len(repo.stored) != 1 || repo.stored[0].TokenHash != "hash:refresh-token" {
			t.Fatalf("expected hashed refresh token to be stored, got %#v", repo.stored)
		}
	})

	t.Run("uses a generic unauthorized error for missing or wrong credentials", func(t *testing.T) {
		cases := []struct {
			name     string
			repo     *fakeRepository
			verifier fakePasswordVerifier
		}{
			{name: "missing email", repo: &fakeRepository{}},
			{name: "wrong password", repo: &fakeRepository{user: activeUser()}, verifier: fakePasswordVerifier{err: errors.New("mismatch")}},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := NewService(tt.repo, tt.verifier, fakeManager{access: "access-token", token: "refresh-token"}, time.Hour)
				_, err := svc.Login(core.LoginRequest{Email: "alice@example.com", Password: "Secret1!"})
				if !errors.Is(err, core.ErrInvalidCredentials) {
					t.Fatalf("expected ErrInvalidCredentials, got %v", err)
				}
			})
		}
	})

	t.Run("rejects inactive users", func(t *testing.T) {
		user := activeUser()
		user.IsActive = false
		svc := NewService(&fakeRepository{user: user}, fakePasswordVerifier{}, fakeManager{access: "access-token", token: "refresh-token"}, time.Hour)
		_, err := svc.Login(core.LoginRequest{Email: "alice@example.com", Password: "Secret1!"})
		if !errors.Is(err, core.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}
