package resolve_auth_user

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetAuthUser(publicID string) (*core.IAMUser, error) {
	var user core.IAMUser
	err := r.db.Preload("Role.Permissions").
		Where("public_id = ?", publicID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
