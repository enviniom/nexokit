package revoke_token

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
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

	t.Run("maps missing token to invalid refresh token sentinel", func(t *testing.T) {
		_, err := repo.GetByHash("hash:missing")
		if !errors.Is(err, core.ErrInvalidRefreshToken) {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("maps zero-row revoke to invalid refresh token", func(t *testing.T) {
		err := repo.Revoke("missing")
		if !errors.Is(err, core.ErrInvalidRefreshToken) {
			t.Fatalf("Revoke() = %v, want invalid refresh token", err)
		}
	})

	sqlDB, err := db.DB()
	if err != nil || sqlDB.Close() != nil {
		t.Fatalf("close db: %v", err)
	}
	assertRefreshPersistenceBoundary(t, repo.Revoke("closed"))
}

func assertRefreshPersistenceBoundary(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != core.CodeRefreshTokenPersistence || appErr.HTTPStatus != http.StatusInternalServerError || appErr.Internal == nil || !errors.Is(err, appErr.Internal) || err == appErr.Internal {
		t.Fatalf("error = %#v, want wrapped refresh-token persistence AppError", err)
	}
}
