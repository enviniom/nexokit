package create_user

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns persistence for the create_user slice.
type Repository interface {
	GetRoleBySlug(slug string) (*iamrole.IAMRole, error)
	ExistsByEmail(email string) (bool, error)
	Create(user *iamuser.IAMUser) error
	GetByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.UserResponse, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new create_user repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetRoleBySlug(slug string) (*iamrole.IAMRole, error) {
	var role iamrole.IAMRole
	if err := r.db.Where("slug = ?", slug).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *GormRepository) ExistsByEmail(email string) (bool, error) {
	_, err := queries.GetUserByEmail(r.db, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *GormRepository) Create(user *iamuser.IAMUser) error {
	if err := r.db.Create(user).Error; err != nil {
		if gormutil.IsUniqueConstraintError(err) {
			return core.ErrUserEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.UserResponse, error) {
	user, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	var response iamuser.UserResponse
	response.FromModel(user)
	return &response, nil
}
