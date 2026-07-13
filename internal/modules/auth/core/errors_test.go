package core

import (
	"errors"
	"net/http"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
)

func TestPersistenceErrorsWrapOriginalCause(t *testing.T) {
	cause := errors.New("database unavailable")
	tests := []struct {
		name string
		make func(error) error
		code apperror.Code
	}{
		{name: "user", make: UserPersistenceError, code: CodeUserPersistence},
		{name: "refresh token", make: RefreshTokenPersistenceError, code: CodeRefreshTokenPersistence},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.make(cause)
			var appErr *apperror.AppError
			if !errors.As(err, &appErr) {
				t.Fatalf("error type = %T, want *apperror.AppError", err)
			}
			if appErr.Code != tt.code || appErr.HTTPStatus != http.StatusInternalServerError {
				t.Fatalf("error = %#v, want code %q and status 500", appErr, tt.code)
			}
			if !errors.Is(err, cause) {
				t.Fatalf("error must preserve cause %v", cause)
			}
		})
	}
}
