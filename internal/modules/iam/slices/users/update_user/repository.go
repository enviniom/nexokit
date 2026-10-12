package update_user

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

// Repository owns persistence for the update_user slice.
type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.IAMUser, error)
	GetRoleBySlug(slug string) (*iamrole.IAMRole, error)
	Update(user *iamuser.IAMUser) error
	Reload(tc tenant.TenantContext, publicID string) (*iamuser.UserResponse, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new update_user repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.IAMUser, error) {
	user, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

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

func (r *GormRepository) Update(user *iamuser.IAMUser) error {
	if err := r.db.Save(user).Error; err != nil {
		if gormutil.IsUniqueConstraintError(err) {
			return core.ErrUserEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GormRepository) Reload(tc tenant.TenantContext, publicID string) (*iamuser.UserResponse, error) {
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
