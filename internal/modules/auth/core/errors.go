package core

import (
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
)

// Module-owned business codes for the auth module.
const (
	CodeInvalidCredentials      apperror.Code = "code:invalid_credentials"
	CodeInvalidRefreshToken     apperror.Code = "code:invalid_refresh_token"
	CodeUserPersistence         apperror.Code = "code:user_persistence_error"
	CodeRefreshTokenPersistence apperror.Code = "code:refresh_token_persistence_error"
)

// Sentinel errors for the auth module.
//
// Both invalid-credentials and invalid-refresh-token are modeled as 401
// Unauthorized so the platform response layer preserves the existing public
// HTTP contract for failed authentication and token rotation/logout flows.
// The public message intentionally reuses the platform's generic Spanish
// "No autorizado" so that callers continue to receive the same body they did
// before the module-owned sentinels were introduced.
var (
	ErrInvalidCredentials  = apperror.Unauthorized(CodeInvalidCredentials, messages.MsgUnauthorized, nil)
	ErrInvalidRefreshToken = apperror.Unauthorized(CodeInvalidRefreshToken, messages.MsgUnauthorized, nil)
)

// UserPersistenceError reports an unexpected user persistence failure while
// retaining its original cause for logging.
func UserPersistenceError(cause error) error {
	return apperror.Internal(CodeUserPersistence, messages.MsgInternalError, cause)
}

// RefreshTokenPersistenceError reports an unexpected refresh-token persistence
// failure while retaining its original cause for logging.
func RefreshTokenPersistenceError(cause error) error {
	return apperror.Internal(CodeRefreshTokenPersistence, messages.MsgInternalError, cause)
}
