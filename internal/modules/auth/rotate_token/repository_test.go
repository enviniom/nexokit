package rotate_token

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepository_GetByHashCreateAndRevoke(t *testing.T) {
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

	refresh := &core.RefreshToken{PublicID: "rt-1", UserID: user.ID, TokenHash: "hash:old", ExpiresAt: time.Now().Add(time.Hour)}
	if err := repo.CreateRefreshToken(refresh); err != nil {
		t.Fatalf("create refresh: %v", err)
	}

	found, err := repo.GetByHash("hash:old")
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if found.User.Role.Slug != "admin" {
		t.Fatalf("expected role preloaded, got %#v", found.User.Role)
	}

	replacement := "hash:new"
	if err := repo.Revoke("hash:old", &replacement); err != nil {
		t.Fatalf("revoke refresh: %v", err)
	}

	stored, err := repo.GetByHash("hash:old")
	if err != nil {
		t.Fatalf("reload by hash: %v", err)
	}
	if stored.RevokedAt == nil || stored.ReplacedByHash == nil || *stored.ReplacedByHash != "hash:new" {
		t.Fatalf("expected revoked token with replacement hash, got %#v", stored)
	}
}
