package permissions

import "github.com/enviniom/nexokit/internal/modules/permissions/core"

// Compatibility aliases preserved for dependent modules during migration.
type Permission = core.Permission
type RolePermission = core.RolePermission

const (
	ActionList              = core.ActionList
	ActionSelect            = core.ActionSelect
	ActionView              = core.ActionView
	ActionCreate            = core.ActionCreate
	ActionUpdate            = core.ActionUpdate
	ActionDelete            = core.ActionDelete
	ActionManage            = core.ActionManage
	ActionChangeRole        = core.ActionChangeRole
	ActionAssignPermissions = core.ActionAssignPermissions
)
