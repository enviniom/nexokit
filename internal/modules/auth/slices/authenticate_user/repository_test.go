package authenticate_user

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

	t.Run("maps missing user to invalid credentials sentinel", func(t *testing.T) {
		_, err := repo.GetByEmail("missing@example.com")
		if !errors.Is(err, core.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("maps missing role to invalid credentials sentinel", func(t *testing.T) {
		userWithoutRole := core.AuthUser{BaseModel: shared.BaseModel{PublicID: "user-without-role"}, Name: "No Role", Email: "no-role@example.com", PasswordHash: "hash", RoleID: role.ID + 1, IsActive: true}
		if err := db.Create(&userWithoutRole).Error; err != nil {
			t.Fatalf("seed user without role: %v", err)
		}

		_, err := repo.GetByEmail(userWithoutRole.Email)
		if !errors.Is(err, core.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestRepository_PersistenceFailuresAreModuleOwned(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
	repo := NewRepository(db)

	_, lookupErr := repo.GetByEmail("alice@example.com")
	assertUserPersistenceError(t, lookupErr)
	createErr := repo.CreateRefreshToken(&core.RefreshToken{TokenHash: "refresh"})
	assertRefreshPersistenceError(t, createErr)
}

func assertUserPersistenceError(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != core.CodeUserPersistence || appErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("error = %#v, want user persistence AppError", err)
	}
	if appErr.Internal == nil || !errors.Is(err, appErr.Internal) || err == appErr.Internal {
		t.Fatalf("error must wrap, not leak, its persistence cause: %#v", err)
	}
}

func assertRefreshPersistenceError(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != core.CodeRefreshTokenPersistence || appErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("error = %#v, want refresh-token persistence AppError", err)
	}
	if appErr.Internal == nil || !errors.Is(err, appErr.Internal) || err == appErr.Internal {
		t.Fatalf("error must wrap, not leak, its persistence cause: %#v", err)
	}
}
