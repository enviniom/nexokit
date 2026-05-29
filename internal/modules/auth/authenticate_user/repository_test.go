package authenticate_user

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepository_GetByEmailAndCreateRefreshToken(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.AuthRole{}, &core.AuthUser{}, &core.RefreshToken{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewRepository(db)

	role := core.AuthRole{BaseModel: shared.BaseModel{PublicID: "role-1"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	user := core.AuthUser{BaseModel: shared.BaseModel{PublicID: "user-1"}, Name: "Alice", Email: "alice@example.com", PasswordHash: "hash", RoleID: role.ID, IsActive: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	found, err := repo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if found.PublicID != "user-1" || found.Role.Name != "Admin" {
		t.Fatalf("expected user with preloaded role, got %#v", found)
	}

	expires := time.Now().Add(time.Hour)
	err = repo.CreateRefreshToken(&core.RefreshToken{PublicID: "rt-1", UserID: found.ID, TokenHash: "hash:refresh", ExpiresAt: expires})
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	var stored core.RefreshToken
	if err := db.Where("token_hash = ?", "hash:refresh").First(&stored).Error; err != nil {
		t.Fatalf("load refresh token: %v", err)
	}
	if stored.UserID != found.ID {
		t.Fatalf("expected user id %d, got %d", found.ID, stored.UserID)
	}
}
