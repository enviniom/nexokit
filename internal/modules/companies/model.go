package companies

import "github.com/enviniom/nexokit/internal/modules/companies/core"

const (
	CompanyStatusActive   = core.CompanyStatusActive
	CompanyStatusInactive = core.CompanyStatusInactive
)

const (
	CompanyDomainStatusActive              = core.CompanyDomainStatusActive
	CompanyDomainStatusInactive            = core.CompanyDomainStatusInactive
	CompanyDomainStatusPendingVerification = core.CompanyDomainStatusPendingVerification
)

const (
	CompanyDomainKindPrimary   = core.CompanyDomainKindPrimary
	CompanyDomainKindAlias     = core.CompanyDomainKindAlias
	CompanyDomainKindTechnical = core.CompanyDomainKindTechnical
)

type Company = core.Company
type CompanyDomain = core.CompanyDomain
