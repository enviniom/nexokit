package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type fakeUserRepository struct {
	byEmail    map[string]*users.User
	byPublicID map[string]*users.User
	err        error
}

func (f *fakeUserRepository) List(page, perPage int) ([]users.User, error) { return nil, nil }
func (f *fakeUserRepository) Count() (int64, error)                        { return 0, nil }
func (f *fakeUserRepository) GetByPublicID(publicID string) (*users.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.byPublicID[publicID]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeUserRepository) GetByEmail(email string) (*users.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeUserRepository) Create(user *users.User) error { return nil }
func (f *fakeUserRepository) Update(user *users.User) error { return nil }
func (f *fakeUserRepository) Delete(publicID string) error  { return nil }

type fakePasswordVerifier struct{ err error }

func (f fakePasswordVerifier) VerifyPassword(password, hash string) error { return f.err }

type fakeTokenIssuer struct {
	access string
	err    error
}

func (f fakeTokenIssuer) IssueAccess(sub, role string, companyID *uint) (string, error) {
	return f.access, f.err
}

type fakeRefreshGenerator struct {
	tokens []string
	idx    int
}

func (f *fakeRefreshGenerator) GenerateRefreshToken() (string, error) {
	tok := f.tokens[f.idx]
	f.idx++
	return tok, nil
}
func (f *fakeRefreshGenerator) HashRefreshToken(token string) string { return "hash:" + token }

type fakeRefreshRepository struct {
	stored  []*RefreshToken
	byHash  map[string]*RefreshToken
	revoked []string
}

func (f *fakeRefreshRepository) Create(refresh *RefreshToken) error {
	f.stored = append(f.stored, refresh)
	if f.byHash == nil {
		f.byHash = map[string]*RefreshToken{}
	}
	f.byHash[refresh.TokenHash] = refresh
	return nil
}
func (f *fakeRefreshRepository) GetByHash(hash string) (*RefreshToken, error) {
	if rt, ok := f.byHash[hash]; ok {
		return rt, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeRefreshRepository) Revoke(hash string, replacedByHash *string) error {
	f.revoked = append(f.revoked, hash)
	if rt, ok := f.byHash[hash]; ok {
		now := time.Now()
		rt.RevokedAt = &now
		rt.ReplacedByHash = replacedByHash
	}
	return nil
}

func testUser(active bool) *users.User {
	return &users.User{
		BaseModel:    shared.BaseModel{ID: 7, PublicID: "user-public-id"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		RoleID:       3,
		Role:         roles.Role{Name: "admin"},
		IsActive:     active,
	}
}

func TestService_Login(t *testing.T) {
	t.Run("issues access and opaque refresh for active user", func(t *testing.T) {
		repo := &fakeRefreshRepository{}
		svc := NewService(
			&fakeUserRepository{byEmail: map[string]*users.User{"alice@example.com": testUser(true)}},
			fakePasswordVerifier{},
			fakeTokenIssuer{access: "access-token"},
			&fakeRefreshGenerator{tokens: []string{"refresh-token"}},
			repo,
			7*24*time.Hour,
		)

		result, err := svc.Login(LoginRequest{Email: "alice@example.com", Password: "Secret1!"})
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
			users    map[string]*users.User
			verifier fakePasswordVerifier
		}{
			{name: "missing email", users: map[string]*users.User{}},
			{name: "wrong password", users: map[string]*users.User{"alice@example.com": testUser(true)}, verifier: fakePasswordVerifier{err: errors.New("mismatch")}},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := NewService(&fakeUserRepository{byEmail: tt.users}, tt.verifier, fakeTokenIssuer{}, &fakeRefreshGenerator{tokens: []string{"refresh"}}, &fakeRefreshRepository{}, time.Hour)
				_, err := svc.Login(LoginRequest{Email: "alice@example.com", Password: "Secret1!"})
				if !errors.Is(err, apperror.ErrUnauthorized) {
					t.Fatalf("expected ErrUnauthorized, got %v", err)
				}
			})
		}
	})

	t.Run("rejects inactive users", func(t *testing.T) {
		svc := NewService(&fakeUserRepository{byEmail: map[string]*users.User{"alice@example.com": testUser(false)}}, fakePasswordVerifier{}, fakeTokenIssuer{}, &fakeRefreshGenerator{tokens: []string{"refresh"}}, &fakeRefreshRepository{}, time.Hour)
		_, err := svc.Login(LoginRequest{Email: "alice@example.com", Password: "Secret1!"})
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestService_Refresh(t *testing.T) {
	t.Run("rotates a valid refresh token and revokes the old hash", func(t *testing.T) {
		old := &RefreshToken{UserID: 7, TokenHash: "hash:old-refresh", ExpiresAt: time.Now().Add(time.Hour), User: *testUser(true)}
		repo := &fakeRefreshRepository{byHash: map[string]*RefreshToken{old.TokenHash: old}}
		svc := NewService(&fakeUserRepository{}, fakePasswordVerifier{}, fakeTokenIssuer{access: "new-access"}, &fakeRefreshGenerator{tokens: []string{"new-refresh"}}, repo, time.Hour)

		result, err := svc.Refresh(RefreshRequest{RefreshToken: "old-refresh"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "new-access" || result.RefreshToken != "new-refresh" {
			t.Fatalf("expected rotated token pair, got %#v", result)
		}
		if len(repo.revoked) != 1 || repo.revoked[0] != "hash:old-refresh" {
			t.Fatalf("expected old token revoked, got %#v", repo.revoked)
		}
		if old.ReplacedByHash == nil || *old.ReplacedByHash != "hash:new-refresh" {
			t.Fatalf("expected replacement hash recorded, got %#v", old.ReplacedByHash)
		}
	})

	t.Run("rejects revoked refresh token", func(t *testing.T) {
		now := time.Now()
		revoked := &RefreshToken{TokenHash: "hash:old-refresh", ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &now, User: *testUser(true)}
		svc := NewService(&fakeUserRepository{}, fakePasswordVerifier{}, fakeTokenIssuer{}, &fakeRefreshGenerator{tokens: []string{"new-refresh"}}, &fakeRefreshRepository{byHash: map[string]*RefreshToken{revoked.TokenHash: revoked}}, time.Hour)

		_, err := svc.Refresh(RefreshRequest{RefreshToken: "old-refresh"})
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestService_Logout(t *testing.T) {
	svc := NewService(&fakeUserRepository{}, fakePasswordVerifier{}, fakeTokenIssuer{}, &fakeRefreshGenerator{tokens: []string{"refresh"}}, &fakeRefreshRepository{byHash: map[string]*RefreshToken{"hash:refresh": {TokenHash: "hash:refresh", ExpiresAt: time.Now().Add(time.Hour)}}}, time.Hour)
	if err := svc.Logout(RefreshRequest{RefreshToken: "refresh"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
