package sync_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryGetBySlugAndAutoAssign(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&core.Permission{}); err != nil { t.Fatal(err) }
	db.Exec("CREATE TABLE roles (id integer primary key, slug text)")
	db.Exec("CREATE TABLE role_permissions (role_id integer, permission_id integer, PRIMARY KEY(role_id,permission_id))")
	db.Exec("INSERT INTO roles(id,slug) VALUES (1,'admin')")

	repo := NewRepository(db)
	p := &core.Permission{Slug: "users.list", Module: "users", Action: "list", Name: "List users", IsSystem: true}
	if err := repo.Create(p); err != nil { t.Fatal(err) }
	if _, err := repo.GetBySlug("users.list"); err != nil { t.Fatal(err) }
	if err := repo.AutoAssignToAdmins(p.ID); err != nil { t.Fatal(err) }
}
