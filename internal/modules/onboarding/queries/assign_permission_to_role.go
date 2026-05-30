package queries

import "gorm.io/gorm"

func AssignPermissionToRole(db *gorm.DB, roleID, permissionID uint) error {
	return db.Table("role_permissions").Create(map[string]any{
		"role_id":       roleID,
		"permission_id": permissionID,
	}).Error
}
