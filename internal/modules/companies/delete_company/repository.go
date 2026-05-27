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
	return queries.GetCompanyByPublicID(r.db, id)
}
func (r *GormRepository) Delete(id string) error {
	return r.db.Where("public_id = ?", id).Delete(&core.Company{}).Error
}
