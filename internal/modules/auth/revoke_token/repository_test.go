package revoke_token

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepository_GetByHashAndRevoke(t *testing.T) {
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
	if err := db.Create(&core.RefreshToken{PublicID: "rt-1", UserID: user.ID, TokenHash: "hash:refresh", ExpiresAt: time.Now().Add(time.Hour)}).Error; err != nil {
		t.Fatalf("seed refresh: %v", err)
	}

	refresh, err := repo.GetByHash("hash:refresh")
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if refresh.User.Role.Slug != "admin" {
		t.Fatalf("expected role preloaded, got %#v", refresh.User.Role)
	}

	if err := repo.Revoke("hash:refresh"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	updated, err := repo.GetByHash("hash:refresh")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if updated.RevokedAt == nil {
		t.Fatalf("expected revoked_at to be set")
	}
}
