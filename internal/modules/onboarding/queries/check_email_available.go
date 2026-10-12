package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/gorm"
)

func CheckEmailAvailable(db *gorm.DB, email string) error {
	// Bootstrap availability follows users.email's unconditional unique constraint:
	// soft-deleted credentials still reserve this exact key; this is not user lookup.
	var count int64
	if err := db.Unscoped().Model(&core.OnboardingUser{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return core.ErrDuplicateAdminEmail
	}
	return nil
}
