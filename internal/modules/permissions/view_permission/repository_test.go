package view_permission

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryUsesSharedQuery(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&core.Permission{}); err != nil { t.Fatal(err) }
	db.Create(&core.Permission{BaseModel: shared.BaseModel{PublicID: "p1"}, Slug: "users.list"})

	p, err := NewRepository(db).GetByPublicID("p1")
	if err != nil || p.Slug != "users.list" {
		t.Fatalf("unexpected repo result: %+v err=%v", p, err)
	}
}
