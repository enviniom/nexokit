package queries

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

// GetByPublicID returns a permission by public ID.
func GetByPublicID(db *gorm.DB, publicID string) (*core.Permission, error) {
	var permission core.Permission
	if err := db.Where("public_id = ?", publicID).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}
