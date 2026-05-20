package goldenmod

import (
	"context"
	"fmt"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// GoldenmodRepository handles persistence for Goldenmod.
type GoldenmodRepository struct {
	db *gorm.DB
}

// NewGoldenmodRepository creates a new repository instance.
func NewGoldenmodRepository(db *gorm.DB) *GoldenmodRepository {
	return &GoldenmodRepository{db: db}
}

// Create persists a new Goldenmod.
func (r *GoldenmodRepository) Create(ctx context.Context, m *Goldenmod) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// FindByPublicID retrieves a Goldenmod by its public ID.
func (r *GoldenmodRepository) FindByPublicID(ctx context.Context, tc tenant.TenantContext, publicID string) (*Goldenmod, error) {
	var m Goldenmod
	q := r.db.WithContext(ctx).Where("public_id = ?", publicID)
	q = tenant.ApplyTenantScope(q, tc)
	if err := q.First(&m).Error; err != nil {
		return nil, fmt.Errorf("goldenmod not found: %w", err)
	}
	return &m, nil
}

// List returns paginated goldenmod.
func (r *GoldenmodRepository) List(ctx context.Context, tc tenant.TenantContext, page, perPage int) ([]Goldenmod, error) {
	var items []Goldenmod
	offset := (page - 1) * perPage
	q := r.db.WithContext(ctx).Limit(perPage).Offset(offset).Order("created_at DESC")
	q = tenant.ApplyTenantScope(q, tc)
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Count returns the total number of goldenmod.
func (r *GoldenmodRepository) Count(ctx context.Context, tc tenant.TenantContext) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&Goldenmod{})
	q = tenant.ApplyTenantScope(q, tc)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Update modifies an existing Goldenmod.
func (r *GoldenmodRepository) Update(ctx context.Context, m *Goldenmod) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete soft-deletes a Goldenmod.
func (r *GoldenmodRepository) Delete(ctx context.Context, tc tenant.TenantContext, publicID string) error {
	q := r.db.WithContext(ctx).Where("public_id = ?", publicID)
	q = tenant.ApplyTenantScope(q, tc)
	return q.Delete(&Goldenmod{}).Error
}
