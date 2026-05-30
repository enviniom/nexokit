package view_permission

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/modules/permissions/queries"
	"gorm.io/gorm"
)

// Repository defines the persistence contract for permissions.
type Repository interface {
	GetByPublicID(publicID string) (*core.Permission, error)
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new view-permission repository.
func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// GetByPublicID returns a permission by public ID.
func (r *GormRepository) GetByPublicID(publicID string) (*core.Permission, error) {
	return queries.GetByPublicID(r.db, publicID)
}
