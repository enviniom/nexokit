package view_company

import "github.com/enviniom/nexokit/internal/modules/companies/core"

type Service interface {
	GetByPublicID(publicID string) (*core.CompanyResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) GetByPublicID(publicID string) (*core.CompanyResponse, error) {
	c, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		return nil, err
	}
	r := &core.CompanyResponse{PublicID: c.PublicID, Name: c.Name, Slug: c.Slug, Status: c.Status, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, CreatedBy: c.CreatedBy, UpdatedBy: c.UpdatedBy}
	if len(c.Domains) > 0 {
		r.Domains = make([]core.CompanyDomainResponse, len(c.Domains))
		for i := range c.Domains {
			r.Domains[i] = core.CompanyDomainResponse{PublicID: c.Domains[i].PublicID, CompanyPublicID: c.PublicID, Domain: c.Domains[i].Domain, Status: c.Domains[i].Status, Kind: c.Domains[i].Kind, RedirectToPrimary: c.Domains[i].RedirectToPrimary, CreatedAt: c.Domains[i].CreatedAt, UpdatedAt: c.Domains[i].UpdatedAt}
		}
	}
	return r, nil
}
