package queries

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

// GetBySlug returns a permission by slug.
func GetBySlug(db *gorm.DB, slug string) (*core.Permission, error) {
	var permission core.Permission
	if err := db.Where("slug = ?", slug).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}
