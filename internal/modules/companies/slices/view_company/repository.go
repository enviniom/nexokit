package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(publicID string) (*core.Company, error)
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(publicID string) (*core.Company, error) {
	var c core.Company
	if err := r.db.Preload("Domains", func(db *gorm.DB) *gorm.DB { return db.Order("kind ASC, domain ASC") }).Where("public_id = ?", publicID).First(&c).Error; err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return &c, nil
}
