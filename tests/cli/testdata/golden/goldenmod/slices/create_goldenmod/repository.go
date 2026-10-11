package create_goldenmod

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

func (r *Repository) Create(ctx context.Context, tc tenant.TenantContext, m *entity.Goldenmod) (*entity.Goldenmod, error) {
	if tc.IsRootScope || tc.CompanyID == 0 {
		return nil, core.ErrTenantRequired
	}
	// An INSERT cannot be constrained by WHERE; assign exclusively from trusted scope.
	m.CompanyID = tc.CompanyID
	q := r.db.WithContext(ctx)
	q = tenant.ApplyTenantScope(q, tc)
	if err := q.Create(m).Error; err != nil {
		return nil, persistenceError(err)
	}
	return m, nil
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
