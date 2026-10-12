package queries

import (
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"gorm.io/gorm"
)

// GetUserByEmail returns the IAM user matching the given email address.
// Returns gorm.ErrRecordNotFound when no match exists.
func GetUserByEmail(db *gorm.DB, email string) (*iamuser.IAMUser, error) {
	var user iamuser.IAMUser
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
