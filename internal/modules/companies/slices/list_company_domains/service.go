package list_company_domains

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
)

type Service interface {
	ListDomains(companyPublicID string) ([]company_domain.CompanyDomainResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) ListDomains(id string) ([]company_domain.CompanyDomainResponse, error) {
	c, err := s.repo.GetByPublicID(id)
	if err != nil {
		return nil, err
	}
	d, err := s.repo.ListDomains(c.ID)
	if err != nil {
		return nil, err
	}
	out := make([]company_domain.CompanyDomainResponse, len(d))
	for i := range d {
		out[i].FromModel(d[i], c.PublicID)
	}
	return out, nil
}
