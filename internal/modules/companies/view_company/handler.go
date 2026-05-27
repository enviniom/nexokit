package view_company

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }
func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	company, err := h.service.GetByPublicID(c.Param("id"))
	if err != nil { response.HandleError(c, err); return }
	response.Success(c, messages.MsgSuccess, company)
}
