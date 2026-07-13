package view_session

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) Handle(c *gin.Context) {
	current, ok := authctx.FromGin(c)
	if !ok || current == nil {
		response.Unauthorized(c, messages.MsgUnauthorized)
		return
	}

	view, err := h.service.View(current)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, messages.MsgSuccess, core.MeResponse{
		AuthUserResponse: core.AuthUserResponse{
			PublicID:  view.PublicID,
			Name:      view.Name,
			Email:     view.Email,
			IsActive:  view.IsActive,
			RoleID:    view.RoleID,
			RoleName:  view.RoleName,
			CompanyID: view.CompanyID,
		},
		RoleSlug:    view.RoleSlug,
		Permissions: view.Permissions,
	})
}
