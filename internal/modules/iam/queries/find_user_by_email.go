package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"gorm.io/gorm"
)

// FindUserByEmail returns one auth user by email.
func FindUserByEmail(db *gorm.DB, email string) (*user.AuthUser, error) {
	var user user.AuthUser
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
