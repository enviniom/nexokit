package update_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"

	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*company.Company, error)
	GetBySlugIncludingDeleted(string) (*company.Company, error)
	Update(*company.Company) error
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(id string) (*company.Company, error) {
	c, err := queries.GetCompanyByPublicID(r.db, id)
	if err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return c, nil
}
func (r *GormRepository) GetBySlugIncludingDeleted(slug string) (*company.Company, error) {
	var c company.Company
	if err := r.db.Unscoped().Where("slug = ?", slug).First(&c).Error; err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return &c, nil
}
func (r *GormRepository) Update(c *company.Company) error {
	result := r.db.Model(&company.Company{}).Where("public_id = ?", c.PublicID).Updates(map[string]any{"name": c.Name, "slug": c.Slug, "status": c.Status})
	if result.Error != nil {
		return queries.MapCompanyError(result.Error)
	}
	if result.RowsAffected == 0 {
		return queries.MapCompanyError(gorm.ErrRecordNotFound)
	}
	return nil
}
