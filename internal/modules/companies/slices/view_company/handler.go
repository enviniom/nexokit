package view_company

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	tc, scoped := tenant.FromGin(c)
	if !scoped {
		response.Forbidden(c, messages.MsgForbidden)
		return
	}
	company, err := h.service.GetByPublicID(tc, c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, company)
}
