package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/gorm"
)

func CheckEmailAvailable(db *gorm.DB, email string) error {
	var count int64
	if err := db.Model(&core.OnboardingUser{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return core.ErrDuplicateAdminEmail
	}
	return nil
}
