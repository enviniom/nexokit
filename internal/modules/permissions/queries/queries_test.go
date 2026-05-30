package queries

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&core.Permission{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	seed := []core.Permission{
		{BaseModel: shared.BaseModel{PublicID: "p3"}, Slug: "users.delete", Module: "users", DisplayOrder: 20},
		{BaseModel: shared.BaseModel{PublicID: "p1"}, Slug: "roles.list", Module: "roles", DisplayOrder: 5},
		{BaseModel: shared.BaseModel{PublicID: "p2"}, Slug: "users.list", Module: "users", DisplayOrder: 10},
	}
	for i := range seed {
		if err := db.Create(&seed[i]).Error; err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	return db
}

func TestGetBySlug(t *testing.T) {
	db := testDB(t)
	p, err := GetBySlug(db, "users.list")
	if err != nil {
		t.Fatalf("get by slug: %v", err)
	}
	if p.PublicID != "p2" {
		t.Fatalf("expected p2, got %s", p.PublicID)
	}
}

func TestListAllSorted(t *testing.T) {
	db := testDB(t)
	items, err := ListAll(db)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(items) != 3 || items[0].Slug != "roles.list" || items[1].Slug != "users.list" {
		t.Fatalf("unexpected order: %+v", items)
	}
}

func TestListPaginated(t *testing.T) {
	db := testDB(t)
	items, err := ListPaginated(db, 2, 1)
	if err != nil {
		t.Fatalf("list paginated: %v", err)
	}
	if len(items) != 1 || items[0].Slug != "users.list" {
		t.Fatalf("unexpected page item: %+v", items)
	}
}
