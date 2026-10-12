package assign_role_to_user

import (
	"context"
	"errors"
	"fmt"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns persistence for the assign_role_to_user slice.
type Repository interface {
	GetUserByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.IAMUser, error)
	GetRoleBySlug(slug string) (*iamrole.IAMRole, error)
	GetAssignableRoleByPublicID(tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error)
	AssignRole(tc tenant.TenantContext, user *iamuser.IAMUser, roleID uint) (*iamuser.UserResponse, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct {
	db    *gorm.DB
	cache cache.Cache
}

// NewRepository creates a new assign_role_to_user repository.
func NewRepository(db *gorm.DB, c cache.Cache) Repository {
	return &GormRepository{db: db, cache: c}
}

func (r *GormRepository) GetUserByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.IAMUser, error) {
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
	role, err := queries.GetRoleBySlug(r.db, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return role, nil
}

func (r *GormRepository) GetAssignableRoleByPublicID(tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error) {
	role, err := queries.GetRoleByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return role, nil
}

func (r *GormRepository) AssignRole(tc tenant.TenantContext, user *iamuser.IAMUser, roleID uint) (*iamuser.UserResponse, error) {
	user.RoleID = roleID
	if err := r.db.Save(user).Error; err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Delete(context.Background(), fmt.Sprintf("rbac:permissions:%s", user.PublicID))
	}

	updated, err := queries.GetUserByPublicID(r.db, tc, user.PublicID)
	if err != nil {
		return nil, err
	}
	var response iamuser.UserResponse
	response.FromModel(updated)
	return &response, nil
}
