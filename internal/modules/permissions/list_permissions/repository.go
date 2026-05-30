package list_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/modules/permissions/queries"
	"gorm.io/gorm"
)

// Repository defines the persistence contract for permissions.
type Repository interface {
	ListAll() ([]core.Permission, error)
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new list-permissions repository.
func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// ListAll returns all permissions in grouped rendering order.
func (r *GormRepository) ListAll() ([]core.Permission, error) {
	return queries.ListAll(r.db)
}
