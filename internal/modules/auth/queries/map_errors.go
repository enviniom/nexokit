package queries

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"gorm.io/gorm"
)

// MapUserError maps user persistence failures to auth domain errors.
func MapUserError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.ErrInvalidCredentials
	}
	return core.UserPersistenceError(err)
}

// MapRefreshTokenError maps refresh-token persistence failures to auth domain errors.
func MapRefreshTokenError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.ErrInvalidRefreshToken
	}
	return core.RefreshTokenPersistenceError(err)
}
