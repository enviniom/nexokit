package create_goldenmod

import (
	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	entity "github.com/enviniom/nexokit/internal/modules/goldenmod/core/goldenmod"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ repo core.CreateRepository }

func NewHandler(repo core.CreateRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok || tc.IsRootScope || tc.CompanyID == 0 {
		response.HandleError(c, core.ErrTenantRequired)
		return
	}
	var req entity.CreateGoldenmodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	m, err := h.repo.Create(c.Request.Context(), tc, &entity.Goldenmod{Name: req.Name})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	var dto entity.GoldenmodResponse
	dto.FromModel(*m)
	response.Created(c, "Goldenmod created", dto)
}
