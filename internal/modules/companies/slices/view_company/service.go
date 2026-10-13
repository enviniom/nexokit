package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*company.CompanyResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) GetByPublicID(tc tenant.TenantContext, publicID string) (*company.CompanyResponse, error) {
	c, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return nil, err
	}
	r := &company.CompanyResponse{}
	r.FromModel(*c, true)
	return r, nil
}
