package view_permission

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetPermissionByPublicID(publicID string) (*core.IAMPermission, error)
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
