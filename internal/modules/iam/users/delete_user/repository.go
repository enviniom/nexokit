package delete_user

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns the persistence for the delete_user slice.
type Repository interface {
	Delete(tc tenant.TenantContext, publicID string) error
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new delete_user repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

// Delete performs a soft-delete on the user matching the given public ID.
// Returns core.ErrNotFound when no row matches.
func (r *GormRepository) Delete(tc tenant.TenantContext, publicID string) error {
	result := tenant.ApplyTenantScope(r.db, tc).Where("public_id = ?", publicID).Delete(&core.IAMUser{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return core.ErrNotFound
	}
	return nil
}
