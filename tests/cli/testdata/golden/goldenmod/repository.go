package goldenmod

import (
	"context"
	"fmt"

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
func (r *GoldenmodRepository) FindByPublicID(ctx context.Context, publicID string) (*Goldenmod, error) {
	var m Goldenmod
	q := r.db.WithContext(ctx).Where("public_id = ?", publicID)
	q = q.Where("company_id = ?", ctx.Value("company_id"))
	if err := q.First(&m).Error; err != nil {
		return nil, fmt.Errorf("goldenmod not found: %w", err)
	}
	return &m, nil
}

// List returns all goldenmod with optional pagination.
func (r *GoldenmodRepository) List(ctx context.Context, limit, offset int) ([]Goldenmod, error) {
	var items []Goldenmod
	q := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC")
	q = q.Where("company_id = ?", ctx.Value("company_id"))
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Update modifies an existing Goldenmod.
func (r *GoldenmodRepository) Update(ctx context.Context, m *Goldenmod) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete soft-deletes a Goldenmod.
func (r *GoldenmodRepository) Delete(ctx context.Context, publicID string) error {
	q := r.db.WithContext(ctx).Where("public_id = ?", publicID)
	q = q.Where("company_id = ?", ctx.Value("company_id"))
	return q.Delete(&Goldenmod{}).Error
}
