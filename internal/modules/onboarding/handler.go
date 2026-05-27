package onboarding

import (
	"errors"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Handler handles onboarding HTTP requests.
type Handler struct {
	service Service
}

// NewHandler creates a new onboarding Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Onboard handles company onboarding HTTP POST request.
func (h *Handler) OnboardCompany(c *gin.Context) {
	var req OnboardCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}

	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	res, err := h.service.Onboard(c.Request.Context(), req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Created(c, messages.MsgCreated, res)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	errs := make(response.ValidationErrors)

	switch {
	case errors.Is(err, ErrDuplicateCompanySlug):
		errs.Add("slug", messages.MsgConflict)
		response.ValidationError(c, errs)
	case errors.Is(err, ErrDuplicateCompanyDomain):
		errs.Add("domain", messages.MsgConflict)
		response.ValidationError(c, errs)
	case errors.Is(err, ErrDuplicateTechnicalDomain):
		errs.Add("generate_technical_domain", messages.MsgConflict)
		response.ValidationError(c, errs)
	case errors.Is(err, ErrMissingPlatformDomain):
		errs.Add("generate_technical_domain", messages.MsgInvalidFormat)
		response.ValidationError(c, errs)
	case errors.Is(err, ErrDuplicateAdminEmail):
		errs.Add("admin_email", messages.MsgConflict)
		response.ValidationError(c, errs)
	default:
		response.HandleError(c, err)
	}
}
