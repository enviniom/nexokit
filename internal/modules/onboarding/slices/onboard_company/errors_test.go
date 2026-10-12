package onboard_company

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
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
			err:        core.ErrDuplicateCompanySlug,
			wantStatus: 422,
			wantCode:   core.CodeDuplicateCompanySlug,
			wantMsg:    "company slug already exists",
		},
		{
			name:       "duplicate company domain",
			err:        core.ErrDuplicateCompanyDomain,
			wantStatus: 422,
			wantCode:   core.CodeDuplicateCompanyDomain,
			wantMsg:    "company domain already exists",
		},
		{
			name:       "duplicate technical domain",
			err:        core.ErrDuplicateTechnicalDomain,
			wantStatus: 422,
			wantCode:   core.CodeDuplicateTechnicalDomain,
			wantMsg:    "company technical domain already exists",
		},
		{
			name:       "missing platform domain",
			err:        core.ErrMissingPlatformDomain,
			wantStatus: 422,
			wantCode:   core.CodeMissingPlatformDomain,
			wantMsg:    "platform domain is required to generate technical domain",
		},
		{
			name:       "duplicate admin email",
			err:        core.ErrDuplicateAdminEmail,
			wantStatus: 422,
			wantCode:   core.CodeDuplicateAdminEmail,
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
