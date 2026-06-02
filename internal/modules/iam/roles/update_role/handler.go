package update_role

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/platform/validator"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Handle(c *gin.Context) {
	var req core.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	tc, ok := tenant.FromGin(c)
	if !ok {
		tc = tenant.NewRoot()
	}
	item, err := h.service.Update(tc, c.Param("id"), req)
	if err != nil {
		if mapped := mapDomainErrorToValidation(err); mapped != nil {
			response.ValidationError(c, mapped)
			return
		}
		response.HandleError(c, mapServiceError(err))
		return
	}
	response.Success(c, messages.MsgSuccess, item)
}

func mapDomainErrorToValidation(err error) validator.ValidationErrors {
	validationErrs := make(validator.ValidationErrors)
	switch {
	case errors.Is(err, core.ErrRoleNameAlreadyExists):
		validationErrs.Add("name", messages.MsgConflict)
	case errors.Is(err, core.ErrRoleSlugAlreadyExists):
		validationErrs.Add("slug", messages.MsgConflict)
	case errors.Is(err, core.ErrReservedRoleIdentity):
		validationErrs.Add("name", messages.MsgConflict)
		validationErrs.Add("slug", messages.MsgConflict)
	default:
		return nil
	}
	return validationErrs
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return apperror.ErrNotFound
	case errors.Is(err, core.ErrRoleProtected):
		return apperror.ErrForbidden
	default:
		return err
	}
}
