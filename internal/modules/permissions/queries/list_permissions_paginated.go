package queries

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

// ListPaginated returns ordered permissions with offset pagination.
func ListPaginated(db *gorm.DB, page, perPage int) ([]core.Permission, error) {
	var permissions []core.Permission
	offset := (page - 1) * perPage
	if err := db.Limit(perPage).Offset(offset).Order("module ASC, display_order ASC, slug ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
