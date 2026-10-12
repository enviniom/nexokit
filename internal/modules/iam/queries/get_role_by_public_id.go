package queries

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

func GetRoleByPublicID(db *gorm.DB, tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error) {
	var role iamrole.IAMRole
	if err := tenant.ApplyTenantScope(db.Preload("Company").Preload("Permissions"), tc).Where("public_id = ?", publicID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
