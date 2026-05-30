package permissions

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewContainerWiresAllSlices(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	c := NewContainer(db, nil, nil)
	if c.ListHandler == nil || c.ViewHandler == nil || c.UpdateHandler == nil || c.Resolver == nil || c.Syncer == nil || c.Catalog == nil {
		t.Fatal("expected all handlers/services wired")
	}
}
