package queries

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"gorm.io/gorm"
)

// MapCompanyError translates company persistence failures at repository boundaries.
func MapCompanyError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.ErrCompanyNotFound
	}
	if gormutil.IsUniqueConstraintError(err) {
		return core.ErrDuplicateCompanySlug
	}
	return core.CompanyPersistenceError(err)
}

// MapCompanyDomainError translates company-domain persistence failures at repository boundaries.
func MapCompanyDomainError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.ErrCompanyDomainNotFound
	}
	if gormutil.IsUniqueConstraintError(err) {
		return core.ErrDuplicateCompanyDomain
	}
	return core.CompanyDomainPersistenceError(err)
}
