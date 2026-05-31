package core

import "github.com/enviniom/nexokit/internal/platform/apperror"

var ErrRoleHasAssignedUsers = apperror.Wrap(apperror.ErrUnprocessable, MsgRoleHasAssignedUsers)
