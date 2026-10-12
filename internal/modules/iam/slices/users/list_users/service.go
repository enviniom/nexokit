package list_users

import (
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Service orchestrates the list_users use case.
type Service interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]iamuser.UserResponse, int64, error)
}

type service struct{ repo Repository }

// NewService creates a list_users service backed by the given repository.
func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) List(tc tenant.TenantContext, params query.ListParams) ([]iamuser.UserResponse, int64, error) {
	items, err := s.repo.List(tc, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(tc, params)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
