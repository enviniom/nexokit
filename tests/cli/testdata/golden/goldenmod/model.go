package goldenmod

import "github.com/enviniom/nexokit/internal/shared"

// Goldenmod is the database model for goldenmod.
type Goldenmod struct {
	shared.BaseModel
	CompanyID uint   `gorm:"index;not null" json:"company_id"`
	Name      string `gorm:"not null" json:"name"`
}
