package users

import "gorm.io/gorm"

// Repository defines the persistence contract for users.
type Repository interface {
	List(page, perPage int) ([]User, error)
	Count() (int64, error)
	GetByPublicID(publicID string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(publicID string) error
	ListPublicIDsByRoleID(roleID uint) ([]string, error)
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM users repository.
func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// List returns paginated users.
func (r *GormRepository) List(page, perPage int) ([]User, error) {
	var users []User
	offset := (page - 1) * perPage
	if err := r.db.Preload("Role").Limit(perPage).Offset(offset).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Count returns the total number of users.
func (r *GormRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetByPublicID returns a user by its public ID.
func (r *GormRepository) GetByPublicID(publicID string) (*User, error) {
	var user User
	if err := r.db.Preload("Role").Where("public_id = ?", publicID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail returns a user by its email.
func (r *GormRepository) GetByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// Create persists a new user.
func (r *GormRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// Update modifies an existing user.
func (r *GormRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// Delete soft-deletes a user by its public ID.
func (r *GormRepository) Delete(publicID string) error {
	return r.db.Where("public_id = ?", publicID).Delete(&User{}).Error
}

// ListPublicIDsByRoleID returns public IDs for users assigned to the role.
func (r *GormRepository) ListPublicIDsByRoleID(roleID uint) ([]string, error) {
	var publicIDs []string
	err := r.db.Model(&User{}).
		Where("role_id = ?", roleID).
		Order("public_id ASC").
		Pluck("public_id", &publicIDs).Error
	return publicIDs, err
}
