package core

import (
	"context"

	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// Each handler consumes only its own persistence capability.
type CreateRepository interface {
	Create(context.Context, tenant.TenantContext, *entity.Goldenmod) (*entity.Goldenmod, error)
}

type GetRepository interface {
	Get(context.Context, tenant.TenantContext, string) (*entity.Goldenmod, error)
}

type ListRepository interface {
	List(context.Context, tenant.TenantContext, query.ListParams) ([]entity.Goldenmod, int64, error)
}

type UpdateRepository interface {
	Update(context.Context, tenant.TenantContext, string, string) (*entity.Goldenmod, error)
}

type DeleteRepository interface {
	Delete(context.Context, tenant.TenantContext, string) error
}
