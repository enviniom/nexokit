package companies

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Resolver struct{ db *gorm.DB }

func NewResolver(db *gorm.DB) *Resolver { return &Resolver{db: db} }

func (r *Resolver) FindByPublicIDOrSlug(value string) (tenant.CompanyRef, error) {
	var company core.Company
	if err := r.db.Where("public_id = ? OR slug = ?", value, value).First(&company).Error; err != nil {
		return tenant.CompanyRef{}, err
	}
	return tenant.CompanyRef{ID: company.ID, Slug: company.Slug}, nil
}

func (r *Resolver) ResolveHost(host string) (tenant.HostResolution, error) {
	var domain core.CompanyDomain
	if err := r.db.Preload("Company").
		Joins("JOIN companies ON companies.id = company_domains.company_id").
		Where("company_domains.domain = ? AND company_domains.status = ? AND companies.status = ?", host, core.CompanyDomainStatusActive, core.CompanyStatusActive).
		First(&domain).Error; err != nil {
		return tenant.HostResolution{}, err
	}

	res := tenant.HostResolution{
		Company:           tenant.CompanyRef{ID: domain.Company.ID, Slug: domain.Company.Slug},
		MatchedDomain:     domain.Domain,
		DomainKind:        domain.Kind,
		RedirectToPrimary: domain.RedirectToPrimary,
	}

	if domain.RedirectToPrimary {
		var primary core.CompanyDomain
		if err := r.db.Where("company_id = ? AND kind = ? AND status = ?", domain.CompanyID, core.CompanyDomainKindPrimary, core.CompanyDomainStatusActive).First(&primary).Error; err == nil {
			res.PrimaryDomain = &primary.Domain
		} else if err != gorm.ErrRecordNotFound {
			return tenant.HostResolution{}, err
		}
	}

	return res, nil
}
