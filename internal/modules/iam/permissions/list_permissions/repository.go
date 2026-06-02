package list_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type Repository interface {
	ListAllPermissions() ([]core.IAMPermission, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) ListAllPermissions() ([]core.IAMPermission, error) {
	var items []core.IAMPermission
	err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&items).Error
	return items, err
}

func toPermissionResponse(p *core.IAMPermission) *core.PermissionResponse {
	return &core.PermissionResponse{
		PublicID:     p.PublicID,
		Slug:         p.Slug,
		Name:         p.Name,
		Module:       p.Module,
		Action:       p.Action,
		Description:  p.Description,
		IsSystem:     p.IsSystem,
		DisplayOrder: p.DisplayOrder,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		CreatedBy:    p.CreatedBy,
		UpdatedBy:    p.UpdatedBy,
	}
}
