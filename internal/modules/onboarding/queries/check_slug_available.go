package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	"gorm.io/gorm"
)

func CheckSlugAvailable(db *gorm.DB, slug string) error {
	// Onboarding must reject exact keys reserved by companies.slug's unconditional
	// unique constraint, including soft-deleted rows; this is not a business read.
	var count int64
	if err := db.Unscoped().Model(&companyentity.OnboardingCompany{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return core.ErrDuplicateCompanySlug
	}
	return nil
}
