package permissions

import "gorm.io/gorm"

// Repository defines the persistence contract for permissions.
type Repository interface {
	List(page, perPage int) ([]Permission, error)
	ListAll() ([]Permission, error)
	Count() (int64, error)
	GetByPublicID(publicID string) (*Permission, error)
	GetBySlug(slug string) (*Permission, error)
	Create(permission *Permission) error
	Update(permission *Permission) error
	Delete(publicID string) error
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM permissions repository.
func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// List returns paginated permissions in grouped rendering order.
func (r *GormRepository) List(page, perPage int) ([]Permission, error) {
	var permissions []Permission
	offset := (page - 1) * perPage
	if err := r.db.Limit(perPage).Offset(offset).Order("module ASC, display_order ASC, slug ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// ListAll returns all permissions in grouped rendering order.
func (r *GormRepository) ListAll() ([]Permission, error) {
	var permissions []Permission
	if err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// Count returns the total number of permissions.
func (r *GormRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&Permission{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetByPublicID returns a permission by public ID.
func (r *GormRepository) GetByPublicID(publicID string) (*Permission, error) {
	var permission Permission
	if err := r.db.Where("public_id = ?", publicID).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetBySlug returns a permission by slug.
func (r *GormRepository) GetBySlug(slug string) (*Permission, error) {
	var permission Permission
	if err := r.db.Where("slug = ?", slug).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// Create persists a new permission.
func (r *GormRepository) Create(permission *Permission) error {
	return r.db.Create(permission).Error
}

// Update modifies an existing permission.
func (r *GormRepository) Update(permission *Permission) error {
	return r.db.Save(permission).Error
}

// Delete soft-deletes a permission by public ID.
func (r *GormRepository) Delete(publicID string) error {
	return r.db.Where("public_id = ?", publicID).Delete(&Permission{}).Error
}
