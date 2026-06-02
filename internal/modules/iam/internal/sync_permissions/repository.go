package sync_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetBySlug(slug string) (*core.IAMPermission, error) {
	var permission core.IAMPermission
	if err := r.db.Where("slug = ?", slug).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *GormRepository) Create(permission *core.IAMPermission) error {
	return r.db.Create(permission).Error
}

func (r *GormRepository) AutoAssignToAdmins(permissionID uint) error {
	return r.db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT id, ? FROM roles WHERE slug = 'admin'
		ON CONFLICT DO NOTHING
	`, permissionID).Error
}
