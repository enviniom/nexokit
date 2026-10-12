package resolve_user_permissions

import "gorm.io/gorm"

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) ListSlugsByUserPublicID(publicID string) ([]string, error) {
	var slugs []string
	err := r.db.Table("permissions").
		Select("permissions.slug").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN users ON users.role_id = role_permissions.role_id").
		Where("users.public_id = ?", publicID).
		Order("permissions.slug ASC").
		Pluck("permissions.slug", &slugs).Error
	return slugs, err
}
