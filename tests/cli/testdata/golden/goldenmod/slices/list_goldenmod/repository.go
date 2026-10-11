package list_goldenmod

import (
	"context"
	"errors"
	"fmt"

	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context, tc tenant.TenantContext, params query.ListParams) ([]entity.Goldenmod, int64, error) {
	// Root-global is an explicit middleware policy, never a missing-context fallback.
	if (!tc.IsRootScope && tc.CompanyID == 0) || (tc.IsRootScope && tc.CompanyID != 0) {
		return nil, 0, core.ErrTenantRequired
	}
	q := r.db.WithContext(ctx).Model(&entity.Goldenmod{})
	q = tenant.ApplyTenantScope(q, tc)
	q = gormutil.ApplySearch(q, params.Search, "name")
	q = gormutil.ApplyDateRange(q, params.Filters, "created_at")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, persistenceError(err)
	}
	// Guard both syntax and the supported public sort fields.
	q = gormutil.ApplySorting(q, params.Sort, "name", "created_at", "updated_at")
	page, perPage := params.Pagination.Page, params.Pagination.PerPage
	if page < 1 {
		page = query.DefaultPage
	}
	if perPage < 1 {
		perPage = query.DefaultPerPage
	}
	if perPage > query.MaxPerPage {
		perPage = query.MaxPerPage
	}
	q = gormutil.ApplyPagination(q, page, perPage)
	items := make([]entity.Goldenmod, 0)
	if err := q.Find(&items).Error; err != nil {
		return nil, 0, persistenceError(err)
	}
	return items, total, nil
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
