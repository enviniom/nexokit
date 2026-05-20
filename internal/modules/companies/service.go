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
var ErrDuplicateSlug = errors.New("company slug already exists")

// Service defines company business operations.
type Service interface {
	List(req ListCompaniesRequest) ([]CompanyResponse, int64, error)
	GetByPublicID(publicID string) (*CompanyResponse, error)
	Create(req CreateCompanyRequest) (*CompanyResponse, error)
	Update(publicID string, req UpdateCompanyRequest) (*CompanyResponse, error)
	Delete(publicID string) error
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
	company := &Company{BaseModel: shared.BaseModel{PublicID: publicID}, Name: req.Name, Slug: slug, Domain: req.Domain, Subdomain: req.Subdomain, Status: status}
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
	company.Domain = req.Domain
	company.Subdomain = req.Subdomain
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
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 20
	}
}

func normalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func toResponse(company *Company) *CompanyResponse {
	return &CompanyResponse{PublicID: company.PublicID, Name: company.Name, Slug: company.Slug, Domain: company.Domain, Subdomain: company.Subdomain, Status: company.Status, CreatedAt: company.CreatedAt, UpdatedAt: company.UpdatedAt, CreatedBy: company.CreatedBy, UpdatedBy: company.UpdatedBy}
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}
