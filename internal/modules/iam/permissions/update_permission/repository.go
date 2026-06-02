package update_permission

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetPermissionByPublicID(publicID string) (*core.IAMPermission, error)
	UpdatePermission(permission *core.IAMPermission) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetPermissionByPublicID(publicID string) (*core.IAMPermission, error) {
	permission, err := queries.GetPermissionByPublicID(r.db, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return permission, nil
}

func (r *GormRepository) UpdatePermission(permission *core.IAMPermission) error {
	err := r.db.Save(permission).Error
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return core.ErrConflict
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "unique failed") {
		return core.ErrConflict
	}

	return err
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
