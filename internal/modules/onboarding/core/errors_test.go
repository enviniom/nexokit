package core

import (
	"errors"
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
			name:       "duplicate company slug",
			err:        ErrDuplicateCompanySlug,
			wantStatus: 422,
			wantCode:   CodeDuplicateCompanySlug,
			wantMsg:    "company slug already exists",
		},
		{
			name:       "duplicate company domain",
			err:        ErrDuplicateCompanyDomain,
			wantStatus: 422,
			wantCode:   CodeDuplicateCompanyDomain,
			wantMsg:    "company domain already exists",
		},
		{
			name:       "duplicate technical domain",
			err:        ErrDuplicateTechnicalDomain,
			wantStatus: 422,
			wantCode:   CodeDuplicateTechnicalDomain,
			wantMsg:    "company technical domain already exists",
		},
		{
			name:       "missing platform domain",
			err:        ErrMissingPlatformDomain,
			wantStatus: 422,
			wantCode:   CodeMissingPlatformDomain,
			wantMsg:    "platform domain is required to generate technical domain",
		},
		{
			name:       "duplicate admin email",
			err:        ErrDuplicateAdminEmail,
			wantStatus: 422,
			wantCode:   CodeDuplicateAdminEmail,
			wantMsg:    "admin email already exists",
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
