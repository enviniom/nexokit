package update_goldenmod

import (
	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ repo core.UpdateRepository }

func NewHandler(repo core.UpdateRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok || tc.IsRootScope || tc.CompanyID == 0 {
		response.HandleError(c, core.ErrTenantRequired)
		return
	}
	var req entity.UpdateGoldenmodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	m, err := h.repo.Update(c.Request.Context(), tc, c.Param("id"), req.Name)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	var dto entity.GoldenmodResponse
	dto.FromModel(*m)
	response.Success(c, "Goldenmod updated", dto)
}
