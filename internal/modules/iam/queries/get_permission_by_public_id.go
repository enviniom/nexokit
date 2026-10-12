package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"

	"gorm.io/gorm"
)

func GetPermissionByPublicID(db *gorm.DB, publicID string) (*permission.IAMPermission, error) {
	var permission permission.IAMPermission
	if err := db.Where("public_id = ?", publicID).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}
