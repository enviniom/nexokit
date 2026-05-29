package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

func TestFindUserByIDWithRole(t *testing.T) {
	db := newAuthQueriesDB(t)

	role := core.AuthRole{
		BaseModel: shared.BaseModel{PublicID: "role_admin"},
		Name:      "Admin",
		Slug:      "admin",
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	user := core.AuthUser{
		BaseModel:    shared.BaseModel{PublicID: "usr_02"},
		Name:         "Bob",
		Email:        "bob@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	t.Run("returns user with role", func(t *testing.T) {
		got, err := FindUserByIDWithRole(db, user.ID)
		if err != nil {
			t.Fatalf("query user by id: %v", err)
		}
		if got.Role.ID != role.ID {
			t.Fatalf("expected role id %d, got %d", role.ID, got.Role.ID)
		}
		if got.Role.Slug != "admin" {
			t.Fatalf("expected role slug admin, got %q", got.Role.Slug)
		}
	})

	t.Run("returns record not found", func(t *testing.T) {
		_, err := FindUserByIDWithRole(db, 999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
		}
	})
}
