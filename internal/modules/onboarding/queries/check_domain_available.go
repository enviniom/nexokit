package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/gorm"
)

func CheckDomainAvailable(db *gorm.DB, domain string, duplicateErr error) error {
	// Onboarding must reject exact keys reserved by company_domains.domain's
	// unconditional unique constraint, including soft-deleted rows, for either kind.
	var count int64
	if err := db.Unscoped().Model(&core.OnboardingCompanyDomain{}).Where("domain = ?", domain).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return duplicateErr
	}
	return nil
}
