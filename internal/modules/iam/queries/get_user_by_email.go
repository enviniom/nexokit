package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

// GetUserByEmail returns the IAM user matching the given email address.
// Returns gorm.ErrRecordNotFound when no match exists.
func GetUserByEmail(db *gorm.DB, email string) (*core.IAMUser, error) {
	var user core.IAMUser
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
