package delete_company

import (
	"errors"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"gorm.io/gorm"
)

type Service interface{ Delete(publicID string) error }
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Delete(id string) error {
	if _, err := s.repo.GetByPublicID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
