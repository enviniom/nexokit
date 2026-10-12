package list_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"

	"gorm.io/gorm"
)

type Repository interface {
	ListAllPermissions() ([]permission.IAMPermission, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) ListAllPermissions() ([]permission.IAMPermission, error) {
	var items []permission.IAMPermission
	err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&items).Error
	return items, err
}
