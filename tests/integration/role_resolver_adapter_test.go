package integration_test

import (
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type roleResolverAdapter struct {
	repo roles.Repository
}

func (r roleResolverAdapter) GetBySlug(slug string) (*roles.Role, error) {
	return r.repo.GetBySlug(tenant.NewRoot(), slug)
}
