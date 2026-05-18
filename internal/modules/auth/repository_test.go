package auth

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthDBFlows_LoginRefreshLogout(t *testing.T) {
	db := newAuthTestDB(t)

	role := roles.Role{
		BaseModel:   shared.BaseModel{PublicID: "role-admin"},
		Name:        "admin",
		Slug:        "admin",
		Description: "Admin role",
		IsSystem:    true,
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	user := users.User{
		BaseModel:    shared.BaseModel{PublicID: "user-alice"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed-password",
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	service := NewService(
		users.NewRepository(db),
		fakePasswordVerifier{},
		fakeTokenIssuer{access: "access-token"},
		&fakeRefreshGenerator{tokens: []string{"login-refresh", "rotated-refresh"}},
		NewRefreshRepository(db),
		time.Hour,
	)

	login, err := service.Login(LoginRequest{Email: "alice@example.com", Password: "Password1"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if login.RefreshToken != "login-refresh" {
		t.Fatalf("expected opaque refresh token returned to client, got %q", login.RefreshToken)
	}

	var stored RefreshToken
	if err := db.Where("token_hash = ?", "hash:login-refresh").First(&stored).Error; err != nil {
		t.Fatalf("expected hashed refresh token stored: %v", err)
	}
	if stored.TokenHash == login.RefreshToken {
		t.Fatal("refresh token was stored in plaintext")
	}

	rotated, err := service.Refresh(RefreshRequest{RefreshToken: "login-refresh"})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if rotated.RefreshToken != "rotated-refresh" {
		t.Fatalf("expected rotated refresh token, got %q", rotated.RefreshToken)
	}

	var old RefreshToken
	if err := db.Where("token_hash = ?", "hash:login-refresh").First(&old).Error; err != nil {
		t.Fatalf("expected old refresh token: %v", err)
	}
	if old.RevokedAt == nil {
		t.Fatal("expected old refresh token to be revoked")
	}
	if old.ReplacedByHash == nil || *old.ReplacedByHash != "hash:rotated-refresh" {
		t.Fatalf("expected replacement hash recorded, got %#v", old.ReplacedByHash)
	}

	var replacement RefreshToken
	if err := db.Where("token_hash = ?", "hash:rotated-refresh").First(&replacement).Error; err != nil {
		t.Fatalf("expected replacement refresh token stored: %v", err)
	}

	if err := service.Logout(RefreshRequest{RefreshToken: "rotated-refresh"}); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	var loggedOut RefreshToken
	if err := db.Where("token_hash = ?", "hash:rotated-refresh").First(&loggedOut).Error; err != nil {
		t.Fatalf("expected logged-out refresh token: %v", err)
	}
	if loggedOut.RevokedAt == nil {
		t.Fatal("expected logout to revoke refresh token")
	}
}

func newAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&roles.Role{}, &users.User{}, &RefreshToken{}); err != nil {
		t.Fatalf("failed to migrate auth test db: %v", err)
	}
	return db
}
