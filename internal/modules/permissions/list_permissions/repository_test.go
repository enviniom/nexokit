package list_permissions

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryListAllOrders(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&core.Permission{}); err != nil { t.Fatal(err) }
	db.Create(&core.Permission{BaseModel: shared.BaseModel{PublicID: "p1"}, Slug: "users.view", Module: "users", DisplayOrder: 20})
	db.Create(&core.Permission{BaseModel: shared.BaseModel{PublicID: "p2"}, Slug: "roles.list", Module: "roles", DisplayOrder: 10})

	items, err := NewRepository(db).ListAll()
	if err != nil || len(items) != 2 || items[0].Slug != "roles.list" {
		t.Fatalf("unexpected items: %+v err=%v", items, err)
	}
}
