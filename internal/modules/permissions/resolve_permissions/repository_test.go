package resolve_permissions

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryListSlugsByUserPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	db.Exec("CREATE TABLE permissions (id integer primary key, slug text)")
	db.Exec("CREATE TABLE role_permissions (role_id integer, permission_id integer)")
	db.Exec("CREATE TABLE users (public_id text, role_id integer)")
	db.Exec("INSERT INTO permissions(id,slug) VALUES (1,'users.list'),(2,'roles.list')")
	db.Exec("INSERT INTO role_permissions(role_id,permission_id) VALUES (10,1),(10,2)")
	db.Exec("INSERT INTO users(public_id,role_id) VALUES ('u1',10)")

	repo := NewRepository(db)
	slugs, err := repo.ListSlugsByUserPublicID("u1")
	if err != nil || len(slugs) != 2 || slugs[0] != "roles.list" {
		t.Fatalf("unexpected slugs: %v err=%v", slugs, err)
	}
}
