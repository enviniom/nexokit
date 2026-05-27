package create_company_domain

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type Service interface {
	CreateDomain(companyPublicID string, req core.CreateCompanyDomainRequest) (*core.CompanyDomainResponse, error)
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) CreateDomain(companyPublicID string, req core.CreateCompanyDomainRequest) (*core.CompanyDomainResponse, error) {
	c, err := s.repo.GetByPublicID(companyPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(req.Domain)), ".")
	if _, err := s.repo.GetDomainByDomain(domain); err == nil {
		return nil, core.ErrDuplicateCompanyDomain
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
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
	d := &core.CompanyDomain{BaseModel: shared.BaseModel{PublicID: pid}, CompanyID: c.ID, Domain: domain, Status: req.Status, Kind: req.Kind, RedirectToPrimary: req.RedirectToPrimary}
	if err := s.repo.CreateDomain(d); err != nil {
		return nil, err
	}
	return &core.CompanyDomainResponse{PublicID: d.PublicID, CompanyPublicID: c.PublicID, Domain: d.Domain, Status: d.Status, Kind: d.Kind, RedirectToPrimary: d.RedirectToPrimary, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
