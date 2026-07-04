package core

import (
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
)

// Module-owned business codes for the IAM module.
const (
	CodeIAMResourceNotFound     apperror.Code = "code:iam_resource_not_found"
	CodeIAMResourceConflict     apperror.Code = "code:iam_resource_conflict"
	CodeIAMForbidden            apperror.Code = "code:iam_forbidden"
	CodeIAMUnauthorized         apperror.Code = "code:iam_unauthorized"
	CodeInvalidPermissionSlug   apperror.Code = "code:invalid_permission_slug"
	CodeRoleNameAlreadyExists   apperror.Code = "code:role_name_already_exists"
	CodeRoleSlugAlreadyExists   apperror.Code = "code:role_slug_already_exists"
	CodeReservedRoleIdentity    apperror.Code = "code:reserved_role_identity"
	CodeRoleHasAssignedUsers    apperror.Code = "code:role_has_assigned_users"
	CodeRoleProtected           apperror.Code = "code:role_protected"
	CodeSystemImmutable         apperror.Code = "code:system_immutable"
	CodeUserEmailAlreadyExists  apperror.Code = "code:user_email_already_exists"
	CodeForbiddenRoleAssignment apperror.Code = "code:forbidden_role_assignment"
	CodeInvalidCompanyScope     apperror.Code = "code:invalid_company_scope"
	CodeForbiddenCompanyScope   apperror.Code = "code:forbidden_company_scope"
)

// Sentinel errors for the IAM module.
var (
	// The public messages for the generic HTTP-category sentinels intentionally
	// reuse the platform's Spanish messages so that callers continue to receive
	// the same response bodies they did before Change 24 removed handler-level
	// remapping to apperror.ErrNotFound, ErrConflict, ErrForbidden,
	// ErrUnauthorized, and ErrBadRequest.
	ErrNotFound                = apperror.NotFound(CodeIAMResourceNotFound, messages.MsgNotFound, nil)
	ErrConflict                = apperror.Conflict(CodeIAMResourceConflict, messages.MsgConflict, nil)
	ErrForbidden               = apperror.Forbidden(CodeIAMForbidden, messages.MsgForbidden, nil)
	ErrUnauthorized            = apperror.Unauthorized(CodeIAMUnauthorized, messages.MsgUnauthorized, nil)
	ErrInvalidPermissionSlug   = apperror.BadRequest(CodeInvalidPermissionSlug, messages.MsgBadRequest, nil)
	ErrRoleNameAlreadyExists   = apperror.Unprocessable(CodeRoleNameAlreadyExists, "role name already exists", nil)
	ErrRoleSlugAlreadyExists   = apperror.Unprocessable(CodeRoleSlugAlreadyExists, "role slug already exists", nil)
	ErrReservedRoleIdentity    = apperror.Unprocessable(CodeReservedRoleIdentity, "reserved role identity", nil)
	ErrRoleHasAssignedUsers    = apperror.Unprocessable(CodeRoleHasAssignedUsers, MsgRoleHasAssignedUsers, nil)
	ErrRoleProtected           = apperror.Forbidden(CodeRoleProtected, "role is protected", nil)
	ErrSystemImmutable         = apperror.Forbidden(CodeSystemImmutable, "system resource is immutable", nil)
	ErrUserEmailAlreadyExists  = apperror.Unprocessable(CodeUserEmailAlreadyExists, "user email already exists", nil)
	ErrForbiddenRoleAssignment = apperror.Forbidden(CodeForbiddenRoleAssignment, "forbidden role assignment", nil)
	ErrInvalidCompanyScope     = apperror.BadRequest(CodeInvalidCompanyScope, messages.MsgBadRequest, nil)
	ErrForbiddenCompanyScope   = apperror.Forbidden(CodeForbiddenCompanyScope, "forbidden company scope", nil)
)
