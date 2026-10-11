package get_goldenmod

import (
	"context"
	"errors"
	"fmt"

	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Get(ctx context.Context, tc tenant.TenantContext, publicID string) (*entity.Goldenmod, error) {
	if tc.IsRootScope || tc.CompanyID == 0 {
		return nil, core.ErrTenantRequired
	}
	var m entity.Goldenmod
	q := r.db.WithContext(ctx).Where("public_id = ?", publicID)
	q = tenant.ApplyTenantScope(q, tc)
	if err := q.First(&m).Error; err != nil {
		return nil, persistenceError(err)
	}
	return &m, nil
}

// Translate known persistence outcomes while retaining their diagnostic cause.
func persistenceError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%w: %w", core.ErrNotFound, err)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return fmt.Errorf("%w: %w", core.ErrConflict, err)
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return fmt.Errorf("%w: %w", core.ErrInvalidReference, err)
	default:
		return err
	}
}
