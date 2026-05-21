package goldenmod

import (
	"context"
	"fmt"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/google/uuid"
)

// GoldenmodService contains business logic for Goldenmod.
type GoldenmodService struct {
	repo *GoldenmodRepository
}

// NewGoldenmodService creates a new service instance.
func NewGoldenmodService(repo *GoldenmodRepository) *GoldenmodService {
	return &GoldenmodService{repo: repo}
}

// Create validates and creates a new Goldenmod.
func (s *GoldenmodService) Create(ctx context.Context, req CreateGoldenmodRequest) (*Goldenmod, error) {
	// TODO: add domain validation before persistence
	m := &Goldenmod{
		Name:      req.Name,
		CompanyID: req.CompanyID,
	}
	m.PublicID = uuid.NewString() // TODO: switch to ULID when identity package is ready
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to create goldenmod: %w", err)
	}
	return m, nil
}

// Get retrieves a Goldenmod by public ID.
func (s *GoldenmodService) Get(ctx context.Context, tc tenant.TenantContext, publicID string) (*Goldenmod, error) {
	return s.repo.FindByPublicID(ctx, tc, publicID)
}

// List returns paginated goldenmod.
func (s *GoldenmodService) List(ctx context.Context, tc tenant.TenantContext, page, perPage int) ([]Goldenmod, int64, error) {
	items, err := s.repo.List(ctx, tc, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, tc)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Update modifies a Goldenmod.
func (s *GoldenmodService) Update(ctx context.Context, tc tenant.TenantContext, publicID string, req UpdateGoldenmodRequest) (*Goldenmod, error) {
	m, err := s.repo.FindByPublicID(ctx, tc, publicID)
	if err != nil {
		return nil, err
	}
	m.Name = req.Name
	m.CompanyID = req.CompanyID
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to update goldenmod: %w", err)
	}
	return m, nil
}

// Delete removes a Goldenmod by public ID.
func (s *GoldenmodService) Delete(ctx context.Context, tc tenant.TenantContext, publicID string) error {
	return s.repo.Delete(ctx, tc, publicID)
}
