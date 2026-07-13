package queries

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"gorm.io/gorm"
)

func TestEntityErrorMappers(t *testing.T) {
	unknown := errors.New("database unavailable")
	for _, tt := range []struct {
		name   string
		mapper func(error) error
		err    error
		want   error
	}{
		{name: "company nil", mapper: MapCompanyError},
		{name: "company not found", mapper: MapCompanyError, err: fmt.Errorf("query: %w", gorm.ErrRecordNotFound), want: core.ErrCompanyNotFound},
		{name: "company unknown", mapper: MapCompanyError, err: unknown},
		{name: "domain nil", mapper: MapCompanyDomainError},
		{name: "domain not found", mapper: MapCompanyDomainError, err: fmt.Errorf("query: %w", gorm.ErrRecordNotFound), want: core.ErrCompanyDomainNotFound},
		{name: "domain unknown", mapper: MapCompanyDomainError, err: unknown},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mapper(tt.err)
			if tt.err == nil && got != nil {
				t.Fatalf("got %v, want nil", got)
			}
			if tt.want != nil && !errors.Is(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if tt.err != nil && tt.want == nil {
				var appErr *apperror.AppError
				if !errors.As(got, &appErr) || appErr.HTTPStatus != http.StatusInternalServerError || !errors.Is(got, tt.err) || got == tt.err {
					t.Fatalf("unknown error must become wrapped 500 AppError, got %#v", got)
				}
			}
		})
	}
}

func TestEntityErrorMappersPreserveExistingAppErrors(t *testing.T) {
	for _, mapper := range []func(error) error{MapCompanyError, MapCompanyDomainError} {
		if got := mapper(core.ErrCompanyNotFound); got != core.ErrCompanyNotFound {
			t.Fatalf("existing AppError must remain unchanged, got %v", got)
		}
	}
}
