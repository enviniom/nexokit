package queries

import (
	permissionentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/permission"
	"gorm.io/gorm"
)

func ListSystemPermissions(db *gorm.DB) ([]permissionentity.OnboardingPermission, error) {
	var permissions []permissionentity.OnboardingPermission
	if err := db.Where("is_system = ?", true).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
