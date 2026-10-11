package goldenmod

import (
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// Goldenmod is a persistence record; expose it through GoldenmodResponse only.
type Goldenmod struct {
	shared.BaseModel
	CompanyID uint   `gorm:"index;not null"`
	Name      string `gorm:"not null"`
}

func (Goldenmod) TableName() string { return "goldenmod" }

// BeforeCreate owns public-ID generation without changing shared model behavior.
func (m *Goldenmod) BeforeCreate(_ *gorm.DB) error {
	if m.PublicID != "" {
		_, err := ulid.ParseStrict(m.PublicID)
		return err
	}
	id, err := identity.Generate()
	if err != nil {
		return err
	}
	m.PublicID = id
	return nil
}
