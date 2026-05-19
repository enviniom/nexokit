package roles

import "gorm.io/gorm"

// Repository defines the persistence contract for roles.
type Repository interface {
	List(page, perPage int) ([]Role, error)
	Count() (int64, error)
	GetByPublicID(publicID string) (*Role, error)
	GetByName(name string) (*Role, error)
	GetBySlug(slug string) (*Role, error)
	Create(role *Role) error
	Update(role *Role) error
	Delete(publicID string) error
	ReplacePermissions(roleID uint, permissionIDs []uint) error
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM roles repository.
func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// List returns paginated roles.
func (r *GormRepository) List(page, perPage int) ([]Role, error) {
	var roles []Role
	offset := (page - 1) * perPage
	if err := r.db.Preload("Permissions").Limit(perPage).Offset(offset).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Count returns the total number of roles.
func (r *GormRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&Role{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetByPublicID returns a role by its public ID.
func (r *GormRepository) GetByPublicID(publicID string) (*Role, error) {
	var role Role
	if err := r.db.Preload("Permissions").Where("public_id = ?", publicID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &role, nil
}

// GetByName returns a role by its name.
func (r *GormRepository) GetByName(name string) (*Role, error) {
	var role Role
	if err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &role, nil
}

// GetBySlug returns a role by its slug.
func (r *GormRepository) GetBySlug(slug string) (*Role, error) {
	var role Role
	if err := r.db.Preload("Permissions").Where("slug = ?", slug).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &role, nil
}

// Create persists a new role.
func (r *GormRepository) Create(role *Role) error {
	return r.db.Create(role).Error
}

// Update modifies an existing role.
func (r *GormRepository) Update(role *Role) error {
	return r.db.Save(role).Error
}

// Delete soft-deletes a role by its public ID.
func (r *GormRepository) Delete(publicID string) error {
	return r.db.Where("public_id = ?", publicID).Delete(&Role{}).Error
}

// ReplacePermissions replaces the role permission join rows exactly.
func (r *GormRepository) ReplacePermissions(roleID uint, permissionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Table("role_permissions").Create(map[string]any{"role_id": roleID, "permission_id": permissionID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
