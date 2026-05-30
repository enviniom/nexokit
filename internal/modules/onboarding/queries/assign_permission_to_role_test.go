package queries

import "testing"

func TestAssignPermissionToRole(t *testing.T) {
	db := newTestDB(t)

	if err := AssignPermissionToRole(db, 7, 11); err != nil {
		t.Fatalf("assign permission to role: %v", err)
	}

	var count int64
	if err := db.Table("role_permissions").Where("role_id = ? AND permission_id = ?", 7, 11).Count(&count).Error; err != nil {
		t.Fatalf("verify role_permissions row: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected role_permissions row count to be 1, got %d", count)
	}
}
