package queries

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

func GetRoleByPublicIDPreloads(db *gorm.DB, tc tenant.TenantContext, publicID string) (*core.IAMRole, error) {
	var role core.IAMRole
	if err := tenant.ApplyTenantScope(db.Preload("Company").Preload("Permissions"), tc).Where("public_id = ?", publicID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
