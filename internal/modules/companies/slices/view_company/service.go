package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
)

type Service interface {
	GetByPublicID(publicID string) (*company.CompanyResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) GetByPublicID(publicID string) (*company.CompanyResponse, error) {
	c, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		return nil, err
	}
	r := &company.CompanyResponse{}
	r.FromModel(*c, true)
	return r, nil
}
