package delete_role

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if err := h.service.Delete(tc, c.Param("id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
