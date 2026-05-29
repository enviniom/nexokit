package queries

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"gorm.io/gorm"
)

// FindUserByEmail returns one auth user by email.
func FindUserByEmail(db *gorm.DB, email string) (*core.AuthUser, error) {
	var user core.AuthUser
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
