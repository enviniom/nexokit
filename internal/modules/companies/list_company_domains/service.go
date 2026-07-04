package list_company_domains

import "github.com/enviniom/nexokit/internal/modules/companies/core"

type Service interface {
	ListDomains(companyPublicID string) ([]core.CompanyDomainResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) ListDomains(id string) ([]core.CompanyDomainResponse, error) {
	c, err := s.repo.GetByPublicID(id)
	if err != nil {
		return nil, err
	}
	d, err := s.repo.ListDomains(c.ID)
	if err != nil {
		return nil, err
	}
	out := make([]core.CompanyDomainResponse, len(d))
	for i := range d {
		out[i] = core.CompanyDomainResponse{PublicID: d[i].PublicID, CompanyPublicID: c.PublicID, Domain: d[i].Domain, Status: d[i].Status, Kind: d[i].Kind, RedirectToPrimary: d[i].RedirectToPrimary, CreatedAt: d[i].CreatedAt, UpdatedAt: d[i].UpdatedAt}
	}
	return out, nil
}
