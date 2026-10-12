package update_permission

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"gorm.io/gorm"
)

type Repository interface {
	GetPermissionByPublicID(publicID string) (*permission.IAMPermission, error)
	UpdatePermission(permission *permission.IAMPermission) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetPermissionByPublicID(publicID string) (*permission.IAMPermission, error) {
	permission, err := queries.GetPermissionByPublicID(r.db, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return permission, nil
}

func (r *GormRepository) UpdatePermission(permission *permission.IAMPermission) error {
	err := r.db.Save(permission).Error
	if err == nil {
		return nil
	}

	if gormutil.IsUniqueConstraintError(err) {
		return core.ErrConflict
	}

	return err
}
