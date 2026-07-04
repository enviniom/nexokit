package onboard_company

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for company onboarding.
type Handler struct {
	service Service
}

// NewHandler creates a new onboarding handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Handle(c *gin.Context) {
	var req core.OnboardCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}

	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	res, err := h.service.Onboard(c.Request.Context(), req)
	if err != nil {
		h.respondOnboardingError(c, err)
		return
	}

	response.Created(c, messages.MsgCreated, res)
}

// respondOnboardingError preserves the original public HTTP contract for the
// known onboarding conflict paths: a 422 Unprocessable Entity response with
// field-keyed validation errors. Any other error is funneled through the
// standard response.HandleError path.
func (h *Handler) respondOnboardingError(c *gin.Context, err error) {
	errs := make(validator.ValidationErrors)

	switch {
	case errors.Is(err, core.ErrDuplicateCompanySlug):
		errs.Add("slug", messages.MsgConflict)
		response.ValidationError(c, errs)
	case errors.Is(err, core.ErrDuplicateCompanyDomain):
		errs.Add("domain", messages.MsgConflict)
		response.ValidationError(c, errs)
	case errors.Is(err, core.ErrDuplicateTechnicalDomain):
		errs.Add("generate_technical_domain", messages.MsgConflict)
		response.ValidationError(c, errs)
	case errors.Is(err, core.ErrMissingPlatformDomain):
		errs.Add("generate_technical_domain", messages.MsgInvalidFormat)
		response.ValidationError(c, errs)
	case errors.Is(err, core.ErrDuplicateAdminEmail):
		errs.Add("admin_email", messages.MsgConflict)
		response.ValidationError(c, errs)
	default:
		response.HandleError(c, err)
	}
}
