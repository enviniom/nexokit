package get_goldenmod

import (
	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ repo core.GetRepository }

func NewHandler(repo core.GetRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok || tc.IsRootScope || tc.CompanyID == 0 {
		response.HandleError(c, core.ErrTenantRequired)
		return
	}
	m, err := h.repo.Get(c.Request.Context(), tc, c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	var dto entity.GoldenmodResponse
	dto.FromModel(*m)
	response.Success(c, "Goldenmod retrieved", dto)
}
