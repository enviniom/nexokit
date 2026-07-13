package core

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
)

func TestSentinels_Status_Code_PublicMessage(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   apperror.Code
		wantMsg    string
	}{
		{
			name:       "company not found",
			err:        ErrCompanyNotFound,
			wantStatus: 404,
			wantCode:   CodeCompanyNotFound,
			wantMsg:    "company not found",
		},
		{
			name:       "company domain not found",
			err:        ErrCompanyDomainNotFound,
			wantStatus: 404,
			wantCode:   CodeCompanyDomainNotFound,
			wantMsg:    "company domain not found",
		},
		{
			name:       "duplicate company domain",
			err:        ErrDuplicateCompanyDomain,
			wantStatus: 409,
			wantCode:   CodeCompanyDomainDuplicate,
			wantMsg:    "company domain already exists",
		},
		{
			name:       "active primary domain exists",
			err:        ErrActivePrimaryDomainExists,
			wantStatus: 422,
			wantCode:   CodePrimaryDomainExists,
			wantMsg:    "company already has an active primary domain",
		},
		{
			name:       "company domain does not belong",
			err:        ErrCompanyDomainDoesNotBelong,
			wantStatus: 404,
			wantCode:   CodeCompanyDomainDoesNotBelong,
			wantMsg:    "company domain does not belong to company",
		},
		{
			name:       "duplicate company slug",
			err:        ErrDuplicateCompanySlug,
			wantStatus: 409,
			wantCode:   CodeCompanySlugDuplicate,
			wantMsg:    "company slug already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apperror.Status(tt.err); got != tt.wantStatus {
				t.Errorf("Status() = %d, want %d", got, tt.wantStatus)
			}

			var ae *apperror.AppError
			if !errors.As(tt.err, &ae) {
				t.Fatalf("expected *apperror.AppError, got %T", tt.err)
			}

			if ae.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", ae.Code, tt.wantCode)
			}

			if !strings.HasPrefix(string(ae.Code), "code:") {
				t.Errorf("Code %q does not start with 'code:' prefix", ae.Code)
			}

			if ae.PublicMessage != tt.wantMsg {
				t.Errorf("PublicMessage = %q, want %q", ae.PublicMessage, tt.wantMsg)
			}
		})
	}
}

func TestPersistenceErrorsWrapOriginalCause(t *testing.T) {
	cause := errors.New("database unavailable")
	for _, tt := range []struct {
		name string
		make func(error) error
		code apperror.Code
	}{
		{name: "company", make: CompanyPersistenceError, code: CodeCompanyPersistence},
		{name: "company domain", make: CompanyDomainPersistenceError, code: CodeCompanyDomainPersistence},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.make(cause)
			var appErr *apperror.AppError
			if !errors.As(err, &appErr) || appErr.Code != tt.code || appErr.HTTPStatus != http.StatusInternalServerError {
				t.Fatalf("expected 500 AppError with %q, got %#v", tt.code, appErr)
			}
			if !errors.Is(err, cause) {
				t.Fatalf("error must preserve cause %v", cause)
			}
		})
	}
}

func TestSentinels_CodeUniqueness(t *testing.T) {
	codes := []apperror.Code{
		CodeCompanyNotFound,
		CodeCompanyDomainNotFound,
		CodeCompanyDomainDuplicate,
		CodePrimaryDomainExists,
		CodeCompanyDomainDoesNotBelong,
		CodeCompanySlugDuplicate,
	}

	seen := make(map[apperror.Code]struct{}, len(codes))
	for _, code := range codes {
		if _, ok := seen[code]; ok {
			t.Errorf("duplicate code %q", code)
		}
		seen[code] = struct{}{}
	}
}
