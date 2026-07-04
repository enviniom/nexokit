package update_company

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
	var req core.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	company, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		if errors.Is(err, core.ErrDuplicateCompanySlug) {
			errs := make(validator.ValidationErrors)
			errs.Add("slug", messages.MsgConflict)
			response.ValidationError(c, errs)
			return
		}
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, company)
}
