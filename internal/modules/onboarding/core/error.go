package core

import "errors"

var (
	ErrDuplicateCompanySlug     = errors.New("company slug already exists")
	ErrDuplicateCompanyDomain   = errors.New("company domain already exists")
	ErrDuplicateTechnicalDomain = errors.New("company technical domain already exists")
	ErrMissingPlatformDomain    = errors.New("platform domain is required to generate technical domain")
	ErrDuplicateAdminEmail      = errors.New("admin email already exists")
)
