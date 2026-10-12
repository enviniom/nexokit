package list_companies

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
)

type Service interface {
	List(req company.ListCompaniesRequest) ([]company.CompanyResponse, int64, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) List(req company.ListCompaniesRequest) ([]company.CompanyResponse, int64, error) {
	if req.ListParams.Pagination.Page < 1 {
		req.ListParams.Pagination.Page = 1
	}
	if req.ListParams.Pagination.PerPage < 1 {
		req.ListParams.Pagination.PerPage = 20
	}
	rows, total, err := s.repo.List(req)
	if err != nil {
		return nil, 0, err
	}
	out := make([]company.CompanyResponse, len(rows))
	for i := range rows {
		out[i].FromModel(rows[i], false)
	}
	return out, total, nil
}
