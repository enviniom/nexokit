package queries

import (
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

func TestFindRefreshTokenByHashWithUser(t *testing.T) {
	db := newAuthQueriesDB(t)

	role := core.AuthRole{BaseModel: shared.BaseModel{PublicID: "role_admin"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	user := core.AuthUser{BaseModel: shared.BaseModel{PublicID: "usr_03"}, Name: "Carla", Email: "carla@example.com", PasswordHash: "hash", RoleID: role.ID, IsActive: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	refresh := core.RefreshToken{PublicID: "rt_03", UserID: user.ID, TokenHash: "hash:refresh", ExpiresAt: time.Now().Add(time.Hour)}
	if err := db.Create(&refresh).Error; err != nil {
		t.Fatalf("seed refresh: %v", err)
	}

	t.Run("returns refresh with user and role", func(t *testing.T) {
		got, err := FindRefreshTokenByHashWithUser(db, "hash:refresh")
		if err != nil {
			t.Fatalf("query refresh by hash: %v", err)
		}
		if got.User.Email != "carla@example.com" || got.User.Role.Slug != "admin" {
			t.Fatalf("expected user+role preloaded, got %#v", got.User)
		}
	})

	t.Run("returns record not found", func(t *testing.T) {
		_, err := FindRefreshTokenByHashWithUser(db, "hash:missing")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
		}
	})
}
