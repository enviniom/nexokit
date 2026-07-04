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
			name:       "not found",
			err:        ErrNotFound,
			wantStatus: 404,
			wantCode:   CodeIAMResourceNotFound,
			wantMsg:    "Recurso no encontrado",
		},
		{
			name:       "conflict",
			err:        ErrConflict,
			wantStatus: 409,
			wantCode:   CodeIAMResourceConflict,
			wantMsg:    "El recurso ya existe",
		},
		{
			name:       "forbidden",
			err:        ErrForbidden,
			wantStatus: 403,
			wantCode:   CodeIAMForbidden,
			wantMsg:    "Acceso denegado",
		},
		{
			name:       "unauthorized",
			err:        ErrUnauthorized,
			wantStatus: 401,
			wantCode:   CodeIAMUnauthorized,
			wantMsg:    "No autorizado",
		},
		{
			name:       "invalid permission slug",
			err:        ErrInvalidPermissionSlug,
			wantStatus: 400,
			wantCode:   CodeInvalidPermissionSlug,
			wantMsg:    "Solicitud inválida",
		},
		{
			name:       "role name already exists",
			err:        ErrRoleNameAlreadyExists,
			wantStatus: 422,
			wantCode:   CodeRoleNameAlreadyExists,
			wantMsg:    "role name already exists",
		},
		{
			name:       "role slug already exists",
			err:        ErrRoleSlugAlreadyExists,
			wantStatus: 422,
			wantCode:   CodeRoleSlugAlreadyExists,
			wantMsg:    "role slug already exists",
		},
		{
			name:       "reserved role identity",
			err:        ErrReservedRoleIdentity,
			wantStatus: 422,
			wantCode:   CodeReservedRoleIdentity,
			wantMsg:    "reserved role identity",
		},
		{
			name:       "role has assigned users",
			err:        ErrRoleHasAssignedUsers,
			wantStatus: 422,
			wantCode:   CodeRoleHasAssignedUsers,
			wantMsg:    MsgRoleHasAssignedUsers,
		},
		{
			name:       "role protected",
			err:        ErrRoleProtected,
			wantStatus: 403,
			wantCode:   CodeRoleProtected,
			wantMsg:    "role is protected",
		},
		{
			name:       "system immutable",
			err:        ErrSystemImmutable,
			wantStatus: 403,
			wantCode:   CodeSystemImmutable,
			wantMsg:    "system resource is immutable",
		},
		{
			name:       "user email already exists",
			err:        ErrUserEmailAlreadyExists,
			wantStatus: 422,
			wantCode:   CodeUserEmailAlreadyExists,
			wantMsg:    "user email already exists",
		},
		{
			name:       "forbidden role assignment",
			err:        ErrForbiddenRoleAssignment,
			wantStatus: 403,
			wantCode:   CodeForbiddenRoleAssignment,
			wantMsg:    "forbidden role assignment",
		},
		{
			name:       "invalid company scope",
			err:        ErrInvalidCompanyScope,
			wantStatus: 400,
			wantCode:   CodeInvalidCompanyScope,
			wantMsg:    "Solicitud inválida",
		},
		{
			name:       "forbidden company scope",
			err:        ErrForbiddenCompanyScope,
			wantStatus: 403,
			wantCode:   CodeForbiddenCompanyScope,
			wantMsg:    "forbidden company scope",
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
		CodeIAMResourceNotFound,
		CodeIAMResourceConflict,
		CodeIAMForbidden,
		CodeIAMUnauthorized,
		CodeInvalidPermissionSlug,
		CodeRoleNameAlreadyExists,
		CodeRoleSlugAlreadyExists,
		CodeReservedRoleIdentity,
		CodeRoleHasAssignedUsers,
		CodeRoleProtected,
		CodeSystemImmutable,
		CodeUserEmailAlreadyExists,
		CodeForbiddenRoleAssignment,
		CodeInvalidCompanyScope,
		CodeForbiddenCompanyScope,
	}

	seen := make(map[apperror.Code]struct{}, len(codes))
	for _, code := range codes {
		if _, ok := seen[code]; ok {
			t.Errorf("duplicate code %q", code)
		}
		seen[code] = struct{}{}
	}
}
