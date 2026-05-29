package authenticate_user

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) Handle(c *gin.Context) {
	var req core.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	result, err := h.service.Login(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, messages.MsgSuccess, result)
}
