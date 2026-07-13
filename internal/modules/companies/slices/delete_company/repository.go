package delete_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*core.Company, error)
	Delete(string) error
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
func (r *GormRepository) Delete(id string) error {
	result := r.db.Where("public_id = ?", id).Delete(&core.Company{})
	if result.Error != nil {
		return queries.MapCompanyError(result.Error)
	}
	if result.RowsAffected == 0 {
		return queries.MapCompanyError(gorm.ErrRecordNotFound)
	}
	return nil
}
