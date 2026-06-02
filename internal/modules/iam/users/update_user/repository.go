package update_user

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns persistence for the update_user slice.
type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMUser, error)
	GetRoleBySlug(slug string) (*core.IAMRole, error)
	Update(user *core.IAMUser) error
	Reload(tc tenant.TenantContext, publicID string) (*core.UserResponse, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new update_user repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMUser, error) {
	user, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *GormRepository) GetRoleBySlug(slug string) (*core.IAMRole, error) {
	var role core.IAMRole
	if err := r.db.Where("slug = ?", slug).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *GormRepository) Update(user *core.IAMUser) error {
	if err := r.db.Save(user).Error; err != nil {
		if isUniqueConstraintError(err) {
			return core.ErrUserEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *GormRepository) Reload(tc tenant.TenantContext, publicID string) (*core.UserResponse, error) {
	user, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return toUserResponse(user), nil
}

func toUserResponse(u *core.IAMUser) *core.UserResponse {
	return &core.UserResponse{
		PublicID:  u.PublicID,
		Name:      u.Name,
		Email:     u.Email,
		IsActive:  u.IsActive,
		RoleID:    u.RoleID,
		RoleName:  u.Role.Name,
		CompanyID: u.CompanyID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		CreatedBy: u.CreatedBy,
		UpdatedBy: u.UpdatedBy,
	}
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}
