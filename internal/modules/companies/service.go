package companies

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// ErrDuplicateSlug marks slug validation failures.
var (
	ErrDuplicateSlug              = errors.New("company slug already exists")
	ErrDuplicateCompanyDomain     = errors.New("company domain already exists")
	ErrActivePrimaryDomainExists  = errors.New("company already has an active primary domain")
	ErrCompanyDomainDoesNotBelong = errors.New("company domain does not belong to company")
)

// Service defines company business operations.
type Service interface {
	List(req ListCompaniesRequest) ([]CompanyResponse, int64, error)
	GetByPublicID(publicID string) (*CompanyResponse, error)
	Create(req CreateCompanyRequest) (*CompanyResponse, error)
	Update(publicID string, req UpdateCompanyRequest) (*CompanyResponse, error)
	Delete(publicID string) error
	ListDomains(companyPublicID string) ([]CompanyDomainResponse, error)
	CreateDomain(companyPublicID string, req CreateCompanyDomainRequest) (*CompanyDomainResponse, error)
	UpdateDomain(companyPublicID, domainPublicID string, req UpdateCompanyDomainRequest) (*CompanyDomainResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a company service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(req ListCompaniesRequest) ([]CompanyResponse, int64, error) {
	normalizeListRequest(&req)
	companies, total, err := s.repo.List(req)
	if err != nil {
		return nil, 0, err
	}
	result := make([]CompanyResponse, len(companies))
	for i := range companies {
		result[i] = *toResponse(&companies[i])
	}
	return result, total, nil
}

func (s *service) GetByPublicID(publicID string) (*CompanyResponse, error) {
	company, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return toResponse(company), nil
}

func (s *service) Create(req CreateCompanyRequest) (*CompanyResponse, error) {
	slug := normalizeSlug(req.Slug)
	if err := s.ensureSlugAvailable(slug, ""); err != nil {
		return nil, err
	}
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = CompanyStatusActive
	}
	company := &Company{BaseModel: shared.BaseModel{PublicID: publicID}, Name: req.Name, Slug: slug, Status: status}
	if err := s.repo.Create(company); err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateSlug
		}
		return nil, err
	}
	return toResponse(company), nil
}

func (s *service) Update(publicID string, req UpdateCompanyRequest) (*CompanyResponse, error) {
	company, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	slug := normalizeSlug(req.Slug)
	if err := s.ensureSlugAvailable(slug, publicID); err != nil {
		return nil, err
	}
	company.Name = req.Name
	company.Slug = slug
	company.Status = req.Status
	if err := s.repo.Update(company); err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateSlug
		}
		return nil, err
	}
	return toResponse(company), nil
}

func (s *service) Delete(publicID string) error {
	if _, err := s.repo.GetByPublicID(publicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	return s.repo.Delete(publicID)
}

func (s *service) ListDomains(companyPublicID string) ([]CompanyDomainResponse, error) {
	company, err := s.repo.GetByPublicID(companyPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	domains, err := s.repo.ListDomains(company.ID)
	if err != nil {
		return nil, err
	}
	result := make([]CompanyDomainResponse, len(domains))
	for i := range domains {
		result[i] = *toDomainResponse(&domains[i], company.PublicID)
	}
	return result, nil
}

func (s *service) CreateDomain(companyPublicID string, req CreateCompanyDomainRequest) (*CompanyDomainResponse, error) {
	company, err := s.repo.GetByPublicID(companyPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	domainName := normalizeCompanyDomain(req.Domain)
	if err := s.ensureDomainAvailable(domainName, ""); err != nil {
		return nil, err
	}
	if err := s.ensureActivePrimaryAllowed(company.ID, req.Kind, req.Status, ""); err != nil {
		return nil, err
	}
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	domain := &CompanyDomain{BaseModel: shared.BaseModel{PublicID: publicID}, CompanyID: company.ID, Domain: domainName, Kind: req.Kind, Status: req.Status, RedirectToPrimary: req.RedirectToPrimary}
	if err := s.repo.CreateDomain(domain); err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateCompanyDomain
		}
		return nil, err
	}
	return toDomainResponse(domain, company.PublicID), nil
}

func (s *service) UpdateDomain(companyPublicID, domainPublicID string, req UpdateCompanyDomainRequest) (*CompanyDomainResponse, error) {
	company, err := s.repo.GetByPublicID(companyPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	domain, err := s.repo.GetDomainByPublicID(domainPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	if domain.CompanyID != company.ID {
		return nil, ErrCompanyDomainDoesNotBelong
	}
	domainName := normalizeCompanyDomain(req.Domain)
	if err := s.ensureDomainAvailable(domainName, domain.PublicID); err != nil {
		return nil, err
	}
	if err := s.ensureActivePrimaryAllowed(company.ID, req.Kind, req.Status, domain.PublicID); err != nil {
		return nil, err
	}
	domain.Domain = domainName
	domain.Kind = req.Kind
	domain.Status = req.Status
	domain.RedirectToPrimary = req.RedirectToPrimary
	if err := s.repo.UpdateDomain(domain); err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateCompanyDomain
		}
		return nil, err
	}
	return toDomainResponse(domain, company.PublicID), nil
}

func (s *service) ensureSlugAvailable(slug, currentPublicID string) error {
	existing, err := s.repo.GetBySlugIncludingDeleted(slug)
	if err == nil {
		if currentPublicID == "" || existing.PublicID != currentPublicID {
			return ErrDuplicateSlug
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func normalizeListRequest(req *ListCompaniesRequest) {
	if req.ListParams.Pagination.Page < 1 {
		req.ListParams.Pagination.Page = 1
	}
	if req.ListParams.Pagination.PerPage < 1 {
		req.ListParams.Pagination.PerPage = 20
	}
}

func (s *service) ensureDomainAvailable(domainName, currentPublicID string) error {
	existing, err := s.repo.GetDomainByDomain(domainName)
	if err == nil {
		if currentPublicID == "" || existing.PublicID != currentPublicID {
			return ErrDuplicateCompanyDomain
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *service) ensureActivePrimaryAllowed(companyID uint, kind, status, currentPublicID string) error {
	if kind != CompanyDomainKindPrimary || status != CompanyDomainStatusActive {
		return nil
	}
	count, err := s.repo.CountActivePrimaryDomains(companyID, currentPublicID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrActivePrimaryDomainExists
	}
	return nil
}

func normalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func normalizeCompanyDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
}

func toResponse(company *Company) *CompanyResponse {
	res := &CompanyResponse{PublicID: company.PublicID, Name: company.Name, Slug: company.Slug, Status: company.Status, CreatedAt: company.CreatedAt, UpdatedAt: company.UpdatedAt, CreatedBy: company.CreatedBy, UpdatedBy: company.UpdatedBy}
	if len(company.Domains) > 0 {
		res.Domains = make([]CompanyDomainResponse, len(company.Domains))
		for i := range company.Domains {
			res.Domains[i] = *toDomainResponse(&company.Domains[i], company.PublicID)
		}
	}
	return res
}

func toDomainResponse(domain *CompanyDomain, companyPublicID string) *CompanyDomainResponse {
	return &CompanyDomainResponse{PublicID: domain.PublicID, CompanyPublicID: companyPublicID, Domain: domain.Domain, Status: domain.Status, Kind: domain.Kind, RedirectToPrimary: domain.RedirectToPrimary, CreatedAt: domain.CreatedAt, UpdatedAt: domain.UpdatedAt}
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}
