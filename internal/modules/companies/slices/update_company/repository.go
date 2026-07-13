package update_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*core.Company, error)
	GetBySlugIncludingDeleted(string) (*core.Company, error)
	Update(*core.Company) error
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(id string) (*core.Company, error) {
	c, err := queries.GetCompanyByPublicID(r.db, id)
	if err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return c, nil
}
func (r *GormRepository) GetBySlugIncludingDeleted(slug string) (*core.Company, error) {
	var c core.Company
	if err := r.db.Unscoped().Where("slug = ?", slug).First(&c).Error; err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return &c, nil
}
func (r *GormRepository) Update(c *core.Company) error {
	result := r.db.Model(&core.Company{}).Where("public_id = ?", c.PublicID).Updates(map[string]any{"name": c.Name, "slug": c.Slug, "status": c.Status})
	if result.Error != nil {
		return queries.MapCompanyError(result.Error)
	}
	if result.RowsAffected == 0 {
		return queries.MapCompanyError(gorm.ErrRecordNotFound)
	}
	return nil
}
