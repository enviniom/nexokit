package create_company_domain

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/platform/shared/string"
	"github.com/enviniom/nexokit/internal/shared"
)

type Service interface {
	CreateDomain(companyPublicID string, req company_domain.CreateCompanyDomainRequest) (*company_domain.CompanyDomainResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) CreateDomain(companyPublicID string, req company_domain.CreateCompanyDomainRequest) (*company_domain.CompanyDomainResponse, error) {
	c, err := s.repo.GetByPublicID(companyPublicID)
	if err != nil {
		return nil, err
	}
	domain := str.NormalizeDomain(req.Domain)
	if _, err := s.repo.GetDomainByDomain(domain); err == nil {
		return nil, core.ErrDuplicateCompanyDomain
	} else if !errors.Is(err, core.ErrCompanyDomainNotFound) {
		return nil, err
	}
	if req.Kind == core.CompanyDomainKindPrimary && req.Status == core.CompanyDomainStatusActive {
		n, err := s.repo.CountActivePrimaryDomains(c.ID, "")
		if err != nil {
			return nil, err
		}
		if n > 0 {
			return nil, core.ErrActivePrimaryDomainExists
		}
	}
	pid, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	d := &company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: pid}, CompanyID: c.ID, Domain: domain, Status: req.Status, Kind: req.Kind, RedirectToPrimary: req.RedirectToPrimary}
	if err := s.repo.CreateDomain(d); err != nil {
		return nil, err
	}
	var dto company_domain.CompanyDomainResponse
	dto.FromModel(*d, c.PublicID)
	return &dto, nil
}
