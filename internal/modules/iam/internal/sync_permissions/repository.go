package sync_permissions

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) FindBySlug(slug string) (*core.IAMPermission, bool, error) {
	var permission core.IAMPermission
	if err := r.db.Where("slug = ?", slug).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &permission, true, nil
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
