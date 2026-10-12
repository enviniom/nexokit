package view_user

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns the persistence for the view_user slice.
type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*iamuser.UserResponse, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new view_user repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

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
