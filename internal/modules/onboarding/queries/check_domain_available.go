package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/gorm"
)

func CheckDomainAvailable(db *gorm.DB, domain string, duplicateErr error) error {
	var count int64
	if err := db.Model(&core.OnboardingCompanyDomain{}).Where("domain = ?", domain).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return duplicateErr
	}
	return nil
}
