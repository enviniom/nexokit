package queries

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

func ExistsRoleByName(db *gorm.DB, tc tenant.TenantContext, name string, excludeRoleID uint) (bool, error) {
	var count int64
	q := tenant.ApplyTenantScope(db.Model(&iamrole.IAMRole{}), tc).Where("name = ?", name)
	if excludeRoleID > 0 {
		q = q.Where("id <> ?", excludeRoleID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
