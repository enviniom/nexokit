package update_company_domain

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	var req core.UpdateCompanyDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	d, err := h.service.UpdateDomain(c.Param("id"), c.Param("domain_id"), req)
	if err != nil {
		if errors.Is(err, core.ErrDuplicateCompanyDomain) {
			v := make(response.ValidationErrors)
			v.Add("domain", messages.MsgConflict)
			response.ValidationError(c, v)
			return
		}
		if errors.Is(err, core.ErrActivePrimaryDomainExists) {
			v := make(response.ValidationErrors)
			v.Add("kind", messages.MsgConflict)
			response.ValidationError(c, v)
			return
		}
		if errors.Is(err, core.ErrCompanyDomainDoesNotBelong) {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, d)
}
