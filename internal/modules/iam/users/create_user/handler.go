package create_user

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

// Handler is the HTTP entry point for the create_user use case.
type Handler struct{ service Service }

// NewHandler creates a create_user handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle creates a new user from a JSON request body.
func (h *Handler) Handle(c *gin.Context) {
	var req core.CreateUserRequest
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
	data, err := h.service.Create(tc, req)
	if err != nil {
		if mapped := mapDomainErrorToValidation(err); mapped != nil {
			response.ValidationError(c, mapped)
			return
		}
		response.HandleError(c, mapServiceError(err))
		return
	}
	response.Created(c, messages.MsgSuccess, data)
}

func mapDomainErrorToValidation(err error) validator.ValidationErrors {
	validationErrs := make(validator.ValidationErrors)
	switch {
	case errors.Is(err, core.ErrUserEmailAlreadyExists):
		validationErrs.Add("email", messages.MsgConflict)
	default:
		return nil
	}
	return validationErrs
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, core.ErrForbiddenRoleAssignment):
		return apperror.ErrForbidden
	case errors.Is(err, core.ErrInvalidCompanyScope):
		return apperror.ErrBadRequest
	case errors.Is(err, core.ErrNotFound):
		return apperror.ErrNotFound
	default:
		return err
	}
}
