package queries

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"gorm.io/gorm"
)

func TestEntityErrorMappers(t *testing.T) {
	unrelated := errors.New("unrelated")

	tests := []struct {
		name   string
		mapper func(error) error
		err    error
		want   error
	}{
		{name: "user nil remains nil", mapper: MapUserError, err: nil, want: nil},
		{name: "user direct record not found maps to invalid credentials", mapper: MapUserError, err: gorm.ErrRecordNotFound, want: core.ErrInvalidCredentials},
		{name: "user wrapped record not found maps to invalid credentials", mapper: MapUserError, err: fmt.Errorf("query: %w", gorm.ErrRecordNotFound), want: core.ErrInvalidCredentials},
		{name: "user unrelated error becomes internal persistence error", mapper: MapUserError, err: unrelated, want: core.UserPersistenceError(unrelated)},
		{name: "refresh token nil remains nil", mapper: MapRefreshTokenError, err: nil, want: nil},
		{name: "refresh token direct record not found maps to invalid refresh token", mapper: MapRefreshTokenError, err: gorm.ErrRecordNotFound, want: core.ErrInvalidRefreshToken},
		{name: "refresh token wrapped record not found maps to invalid refresh token", mapper: MapRefreshTokenError, err: fmt.Errorf("query: %w", gorm.ErrRecordNotFound), want: core.ErrInvalidRefreshToken},
		{name: "refresh token unrelated error becomes internal persistence error", mapper: MapRefreshTokenError, err: unrelated, want: core.RefreshTokenPersistenceError(unrelated)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mapper(tt.err)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("mapper() = %v, want nil", got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("mapper() = %v, want errors.Is(_, %v)", got, tt.want)
			}
			var appErr *apperror.AppError
			if !errors.As(got, &appErr) {
				t.Fatalf("mapper() must return an AppError, got %T", got)
			}
			if errors.Is(tt.err, gorm.ErrRecordNotFound) {
				if appErr.HTTPStatus != http.StatusUnauthorized {
					t.Fatalf("mapper() status = %d, want %d", appErr.HTTPStatus, http.StatusUnauthorized)
				}
			} else {
				if appErr.HTTPStatus != http.StatusInternalServerError {
					t.Fatalf("mapper() status = %d, want %d", appErr.HTTPStatus, http.StatusInternalServerError)
				}
				if !errors.Is(got, tt.err) {
					t.Fatalf("mapper() must preserve cause %v", tt.err)
				}
				if got == tt.err {
					t.Fatal("mapper() leaked the raw persistence error")
				}
			}
		})
	}
}
