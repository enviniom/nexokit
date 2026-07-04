package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
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
			name:       "invalid credentials",
			err:        ErrInvalidCredentials,
			wantStatus: 401,
			wantCode:   CodeInvalidCredentials,
			wantMsg:    messages.MsgUnauthorized,
		},
		{
			name:       "invalid refresh token",
			err:        ErrInvalidRefreshToken,
			wantStatus: 401,
			wantCode:   CodeInvalidRefreshToken,
			wantMsg:    messages.MsgUnauthorized,
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

func TestSentinels_CodeUniqueness(t *testing.T) {
	codes := []apperror.Code{
		CodeInvalidCredentials,
		CodeInvalidRefreshToken,
	}

	seen := make(map[apperror.Code]struct{}, len(codes))
	for _, code := range codes {
		if _, ok := seen[code]; ok {
			t.Errorf("duplicate code %q", code)
		}
		seen[code] = struct{}{}
	}
}
