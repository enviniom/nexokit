package update_user

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/platform/validator"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the update_user use case.
type Handler struct{ service Service }

// NewHandler creates an update_user handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle updates an existing user from a JSON request body.
func (h *Handler) Handle(c *gin.Context) {
	var req core.UpdateUserRequest
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

	data, err := h.service.Update(tc, c.Param("id"), authctx.PublicIDFromGin(c), req)
	if err != nil {
		if mapped := mapDomainErrorToValidation(err); mapped != nil {
			response.ValidationError(c, mapped)
			return
		}
		response.HandleError(c, err)
		return
	}

	response.Success(c, messages.MsgSuccess, data)
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
