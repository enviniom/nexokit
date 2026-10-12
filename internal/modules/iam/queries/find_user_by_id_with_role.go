package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"gorm.io/gorm"
)

// FindUserByIDWithRole returns one auth user by id with role preloaded.
func FindUserByIDWithRole(db *gorm.DB, id uint) (*user.AuthUser, error) {
	var user user.AuthUser
	if err := db.Preload("Role").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
