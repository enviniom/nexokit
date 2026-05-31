package create_company_domain

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	var req core.CreateCompanyDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	d, err := h.service.CreateDomain(c.Param("id"), req)
	if err != nil {
		if errors.Is(err, core.ErrDuplicateCompanyDomain) {
			v := make(validator.ValidationErrors)
			v.Add("domain", messages.MsgConflict)
			response.ValidationError(c, v)
			return
		}
		if errors.Is(err, core.ErrActivePrimaryDomainExists) {
			v := make(validator.ValidationErrors)
			v.Add("kind", messages.MsgConflict)
			response.ValidationError(c, v)
			return
		}
		response.HandleError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, d)
}
