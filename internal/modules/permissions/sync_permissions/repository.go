package sync_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/modules/permissions/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetBySlug(slug string) (*core.Permission, error)
	Create(permission *core.Permission) error
	AutoAssignToAdmins(permissionID uint) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetBySlug(slug string) (*core.Permission, error) {
	return queries.GetBySlug(r.db, slug)
}

func (r *GormRepository) Create(permission *core.Permission) error {
	return r.db.Create(permission).Error
}

func (r *GormRepository) AutoAssignToAdmins(permissionID uint) error {
	return r.db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT id, ? FROM roles WHERE slug = 'admin'
		ON CONFLICT DO NOTHING
	`, permissionID).Error
}
