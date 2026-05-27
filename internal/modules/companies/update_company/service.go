package update_company

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"gorm.io/gorm"
)

var ErrDuplicateSlug = errors.New("company slug already exists")

type Service interface {
	Update(publicID string, req core.UpdateCompanyRequest) (*core.CompanyResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Update(publicID string, req core.UpdateCompanyRequest) (*core.CompanyResponse, error) {
	c, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if ex, err := s.repo.GetBySlugIncludingDeleted(slug); err == nil && ex.PublicID != publicID {
		return nil, ErrDuplicateSlug
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	c.Name, c.Slug, c.Status = req.Name, slug, req.Status
	if err := s.repo.Update(c); err != nil {
		return nil, err
	}
	return &core.CompanyResponse{PublicID: c.PublicID, Name: c.Name, Slug: c.Slug, Status: c.Status, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, CreatedBy: c.CreatedBy, UpdatedBy: c.UpdatedBy}, nil
}
