package core

import (
	"errors"
)

var (
	ErrNotFound                = errors.New("iam resource not found")
	ErrConflict                = errors.New("iam resource conflict")
	ErrForbidden               = errors.New("iam forbidden")
	ErrUnauthorized            = errors.New("iam unauthorized")
	ErrInvalidPermissionSlug   = errors.New("iam invalid permission slug")
	ErrRoleNameAlreadyExists   = errors.New("iam role name already exists")
	ErrRoleSlugAlreadyExists   = errors.New("iam role slug already exists")
	ErrReservedRoleIdentity    = errors.New("iam reserved role identity")
	ErrRoleHasAssignedUsers    = errors.New("iam role has assigned users")
	ErrRoleProtected           = errors.New("iam role protected")
	ErrSystemImmutable         = errors.New("system resource is immutable")
	ErrUserEmailAlreadyExists  = errors.New("iam user email already exists")
	ErrForbiddenRoleAssignment = errors.New("iam forbidden role assignment")
	ErrInvalidCompanyScope     = errors.New("iam invalid company scope")
)
