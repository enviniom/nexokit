package update_permission

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryGetByPublicIDAndUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&core.Permission{}); err != nil { t.Fatal(err) }
	db.Create(&core.Permission{BaseModel: shared.BaseModel{PublicID: "p1"}, Slug: "users.list", Name: "Old"})

	repo := NewRepository(db)
	p, err := repo.GetByPublicID("p1")
	if err != nil { t.Fatal(err) }
	p.Name = "New"
	if err := repo.Update(p); err != nil { t.Fatal(err) }
}
