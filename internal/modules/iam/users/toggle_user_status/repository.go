package toggle_user_status

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns persistence for the toggle_user_status slice.
type Repository interface {
	ToggleStatus(tc tenant.TenantContext, publicID string, isActive bool) (*core.UserResponse, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new toggle_user_status repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) ToggleStatus(tc tenant.TenantContext, publicID string, isActive bool) (*core.UserResponse, error) {
	user, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	user.IsActive = isActive
	if err := r.db.Save(user).Error; err != nil {
		return nil, err
	}

	updated, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		return nil, err
	}
	return toUserResponse(updated), nil
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
