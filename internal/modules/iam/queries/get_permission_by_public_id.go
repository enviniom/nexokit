package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

func GetPermissionByPublicID(db *gorm.DB, publicID string) (*core.IAMPermission, error) {
	var permission core.IAMPermission
	if err := db.Where("public_id = ?", publicID).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}
