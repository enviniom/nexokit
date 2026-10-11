package list_goldenmod

import (
	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ repo core.ListRepository }

func NewHandler(repo core.ListRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok || (!tc.IsRootScope && tc.CompanyID == 0) || (tc.IsRootScope && tc.CompanyID != 0) {
		response.HandleError(c, core.ErrTenantRequired)
		return
	}
	params := query.ListFromGin(c)
	items, total, err := h.repo.List(c.Request.Context(), tc, params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	dtos := make([]entity.GoldenmodResponse, len(items))
	for i, m := range items {
		dtos[i].FromModel(m)
	}
	response.PaginatedWithFilters(c, "Goldenmod list retrieved", dtos, params, total)
}
