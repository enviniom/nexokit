package update_company

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/shared/string"
)

type Service interface {
	Update(publicID string, req company.UpdateCompanyRequest) (*company.CompanyResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Update(publicID string, req company.UpdateCompanyRequest) (*company.CompanyResponse, error) {
	c, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		return nil, err
	}
	slug := str.NormalizeSlug(req.Slug)
	if ex, err := s.repo.GetBySlugIncludingDeleted(slug); err == nil && ex.PublicID != publicID {
		return nil, core.ErrDuplicateCompanySlug
	} else if err != nil && !errors.Is(err, core.ErrCompanyNotFound) {
		return nil, err
	}
	c.Name, c.Slug, c.Status = req.Name, slug, req.Status
	if err := s.repo.Update(c); err != nil {
		return nil, err
	}
	var dto company.CompanyResponse
	dto.FromModel(*c, false)
	return &dto, nil
}
