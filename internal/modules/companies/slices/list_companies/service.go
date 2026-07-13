package list_companies

import "github.com/enviniom/nexokit/internal/modules/companies/core"

type Service interface {
	List(req core.ListCompaniesRequest) ([]core.CompanyResponse, int64, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) List(req core.ListCompaniesRequest) ([]core.CompanyResponse, int64, error) {
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
	out := make([]core.CompanyResponse, len(rows))
	for i := range rows {
		out[i] = core.CompanyResponse{PublicID: rows[i].PublicID, Name: rows[i].Name, Slug: rows[i].Slug, Status: rows[i].Status, CreatedAt: rows[i].CreatedAt, UpdatedAt: rows[i].UpdatedAt, CreatedBy: rows[i].CreatedBy, UpdatedBy: rows[i].UpdatedBy}
	}
	return out, total, nil
}
