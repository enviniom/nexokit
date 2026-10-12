package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/gorm"
)

func CheckSlugAvailable(db *gorm.DB, slug string) error {
	// Onboarding must reject exact keys reserved by companies.slug's unconditional
	// unique constraint, including soft-deleted rows; this is not a business read.
	var count int64
	if err := db.Unscoped().Model(&core.OnboardingCompany{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return core.ErrDuplicateCompanySlug
	}
	return nil
}
