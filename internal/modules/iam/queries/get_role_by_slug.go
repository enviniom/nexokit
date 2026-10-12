package queries

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"gorm.io/gorm"
)

// GetRoleBySlug fetches a role by its unique slug. Not tenant-scoped because
// it is primarily used for global/reserved role checks (e.g. root).
func GetRoleBySlug(db *gorm.DB, slug string) (*iamrole.IAMRole, error) {
	var role iamrole.IAMRole
	if err := db.Where("slug = ?", slug).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
