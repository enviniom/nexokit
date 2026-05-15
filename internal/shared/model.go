package shared

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel is the GORM persistence layer. Do not add json tags here.
// Define DTOs in each module to control the HTTP contract.
type BaseModel struct {
	ID        uint           `gorm:"primaryKey"`
	PublicID  string         `gorm:"type:char(26);uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy *uint          `gorm:"index"`
	UpdatedBy *uint          `gorm:"index"`
}

// BaseModelSimple provides the standard fields without audit tracking.
type BaseModelSimple struct {
	ID        uint           `gorm:"primaryKey"`
	PublicID  string         `gorm:"type:char(26);uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
