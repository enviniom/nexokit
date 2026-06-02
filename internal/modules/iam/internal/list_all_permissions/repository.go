package list_all_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) ListAllPermissions() ([]core.IAMPermission, error) {
	var items []core.IAMPermission
	err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&items).Error
	return items, err
}
