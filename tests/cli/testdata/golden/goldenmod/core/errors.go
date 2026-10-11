package core

import (
	"github.com/enviniom/nexokit/internal/platform/apperror"
)

var (
	ErrNotFound         = apperror.NotFound("goldenmod.not_found", "Goldenmod not found", nil)
	ErrConflict         = apperror.Conflict("goldenmod.conflict", "Goldenmod already exists", nil)
	ErrInvalidReference = apperror.BadRequest("goldenmod.invalid_reference", "Invalid goldenmod reference", nil)
	ErrTenantRequired   = apperror.BadRequest("goldenmod.tenant_required", "Valid company scope required", nil)
)
