package roles

import "github.com/enviniom/nexokit/internal/shared"

// Role represents a user role in the system.
type Role struct {
	shared.BaseModel
	Name        string `gorm:"uniqueIndex;not null"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Description string
	IsSystem    bool   `gorm:"not null;default:false"`
}
