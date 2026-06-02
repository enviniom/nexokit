package assign_permissions_to_role

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	var req core.AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	tc, ok := tenant.FromGin(c)
	if !ok {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	item, err := h.service.Assign(tc, c.Param("id"), req, permissionSlugsFromContext(c))
	if err != nil {
		response.HandleError(c, mapServiceError(err))
		return
	}
	response.Success(c, messages.MsgSuccess, item)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return apperror.ErrNotFound
	case errors.Is(err, core.ErrRoleProtected), errors.Is(err, core.ErrSystemImmutable):
		return apperror.ErrForbidden
	case errors.Is(err, core.ErrInvalidPermissionSlug):
		return apperror.ErrBadRequest
	default:
		return err
	}
}

func permissionSlugsFromContext(c *gin.Context) []string {
	user, ok := authctx.FromGin(c)
	if !ok || user == nil {
		return nil
	}
	return user.Permissions
}
