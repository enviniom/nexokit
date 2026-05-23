package permissions

import (
	"fmt"
	"strings"
)

const (
	ModuleUsers       = "users"
	ModuleRoles       = "roles"
	ModuleCompanies   = "companies"
	ModuleSettings    = "settings"
	ModuleAuth        = "auth"
	ModulePermissions = "permissions"
)

const (
	ActionList              = "list"
	ActionSelect            = "select"
	ActionView              = "view"
	ActionCreate            = "create"
	ActionUpdate            = "update"
	ActionDelete            = "delete"
	ActionManage            = "manage"
	ActionChangeRole        = "change_role"
	ActionAssignPermissions = "assign_permissions"
)

// Format compiles module and action into a standard permission slug string.
func Format(module, action string) string {
	return fmt.Sprintf("%s.%s", module, action)
}

// HumanizeName translates a module and action into a clean, readable name.
func HumanizeName(module, action string) string {
	actionName := action
	switch action {
	case ActionList:
		actionName = "List"
	case ActionSelect:
		actionName = "Select"
	case ActionView:
		actionName = "View"
	case ActionCreate:
		actionName = "Create"
	case ActionUpdate:
		actionName = "Update"
	case ActionDelete:
		actionName = "Delete"
	case ActionManage:
		actionName = "Manage"
	case ActionChangeRole:
		actionName = "Change role"
	case ActionAssignPermissions:
		actionName = "Assign permissions"
	default:
		if len(action) > 0 {
			actionName = strings.ToUpper(action[:1]) + action[1:]
		}
	}
	return fmt.Sprintf("%s %s", actionName, module)
}

// HumanizeDescription generates a friendly default description for a permission.
func HumanizeDescription(module, action string) string {
	actionDesc := action
	switch action {
	case ActionList:
		actionDesc = "listing"
	case ActionSelect:
		actionDesc = "selecting"
	case ActionView:
		actionDesc = "viewing"
	case ActionCreate:
		actionDesc = "creating"
	case ActionUpdate:
		actionDesc = "updating"
	case ActionDelete:
		actionDesc = "deleting"
	case ActionManage:
		actionDesc = "managing"
	case ActionChangeRole:
		actionDesc = "changing role of"
	case ActionAssignPermissions:
		actionDesc = "assigning permissions for"
	}
	return fmt.Sprintf("Allows %s %s", actionDesc, module)
}

// DefaultDisplayOrder assigns standard cosmetic visual weights to actions.
func DefaultDisplayOrder(action string) int {
	switch action {
	case ActionList, ActionManage:
		return 10
	case ActionView:
		return 20
	case ActionCreate:
		return 30
	case ActionUpdate:
		return 40
	case ActionDelete:
		return 50
	case ActionChangeRole, ActionAssignPermissions:
		return 60
	case ActionSelect:
		return 70
	default:
		return 100
	}
}
