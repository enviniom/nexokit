package queries

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

// ListAll returns all permissions sorted by module, display_order, slug.
func ListAll(db *gorm.DB) ([]core.Permission, error) {
	var permissions []core.Permission
	if err := db.Order("module ASC, display_order ASC, slug ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
