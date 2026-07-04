package update_company_domain

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/shared/string"
)

type Service interface {
	UpdateDomain(companyPublicID, domainPublicID string, req core.UpdateCompanyDomainRequest) (*core.CompanyDomainResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) UpdateDomain(companyPublicID, domainPublicID string, req core.UpdateCompanyDomainRequest) (*core.CompanyDomainResponse, error) {
	c, err := s.repo.GetByPublicID(companyPublicID)
	if err != nil {
		return nil, err
	}
	d, err := s.repo.GetDomainByPublicID(domainPublicID)
	if err != nil {
		return nil, err
	}
	if d.CompanyID != c.ID {
		return nil, core.ErrCompanyDomainDoesNotBelong
	}
	domain := str.NormalizeDomain(req.Domain)
	if ex, err := s.repo.GetDomainByDomain(domain); err == nil && ex.PublicID != d.PublicID {
		return nil, core.ErrDuplicateCompanyDomain
	} else if err != nil && !errors.Is(err, core.ErrCompanyDomainNotFound) {
		return nil, err
	}
	if req.Kind == core.CompanyDomainKindPrimary && req.Status == core.CompanyDomainStatusActive {
		n, err := s.repo.CountActivePrimaryDomains(c.ID, d.PublicID)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			return nil, core.ErrActivePrimaryDomainExists
		}
	}
	d.Domain, d.Kind, d.Status, d.RedirectToPrimary = domain, req.Kind, req.Status, req.RedirectToPrimary
	if err := s.repo.UpdateDomain(d); err != nil {
		return nil, err
	}
	return &core.CompanyDomainResponse{PublicID: d.PublicID, CompanyPublicID: c.PublicID, Domain: d.Domain, Status: d.Status, Kind: d.Kind, RedirectToPrimary: d.RedirectToPrimary, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
