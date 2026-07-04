package update_company

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrCompanyNotFound
		}
		return nil, err
	}
	return c, nil
}
func (r *GormRepository) GetBySlugIncludingDeleted(slug string) (*core.Company, error) {
	var c core.Company
	if err := r.db.Unscoped().Where("slug = ?", slug).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrCompanyNotFound
		}
		return nil, err
	}
	return &c, nil
}
func (r *GormRepository) Update(c *core.Company) error {
	if err := r.db.Save(c).Error; err != nil {
		if gormutil.IsUniqueConstraintError(err) {
			return core.ErrDuplicateCompanySlug
		}
		return err
	}
	return nil
}
