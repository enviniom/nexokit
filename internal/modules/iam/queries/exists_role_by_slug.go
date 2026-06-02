package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

func ExistsRoleBySlug(db *gorm.DB, tc tenant.TenantContext, slug string, excludeRoleID uint) (bool, error) {
	var count int64
	q := tenant.ApplyTenantScope(db.Model(&core.IAMRole{}), tc).Where("slug = ?", slug)
	if excludeRoleID > 0 {
		q = q.Where("id <> ?", excludeRoleID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
