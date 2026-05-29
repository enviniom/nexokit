package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFindUserByEmail(t *testing.T) {
	db := newAuthQueriesDB(t)

	seed := core.AuthUser{
		BaseModel:    shared.BaseModel{PublicID: "usr_01"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       1,
		IsActive:     true,
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	t.Run("returns matched user", func(t *testing.T) {
		got, err := FindUserByEmail(db, "alice@example.com")
		if err != nil {
			t.Fatalf("query user: %v", err)
		}
		if got.Email != "alice@example.com" {
			t.Fatalf("expected alice@example.com, got %q", got.Email)
		}
	})

	t.Run("returns record not found", func(t *testing.T) {
		_, err := FindUserByEmail(db, "missing@example.com")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
		}
	})
}

func newAuthQueriesDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&core.AuthRole{}, &core.AuthUser{}, &core.RefreshToken{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	return db
}
