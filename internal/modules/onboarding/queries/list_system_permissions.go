package queries

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"gorm.io/gorm"
)

func ListSystemPermissions(db *gorm.DB) ([]core.OnboardingPermission, error) {
	var permissions []core.OnboardingPermission
	if err := db.Where("is_system = ?", true).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
