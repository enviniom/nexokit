package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

func GetUserByPublicID(db *gorm.DB, tc tenant.TenantContext, publicID string) (*core.IAMUser, error) {
	var user core.IAMUser
	if err := tenant.ApplyTenantScope(db.Preload("Role"), tc).Where("public_id = ?", publicID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
