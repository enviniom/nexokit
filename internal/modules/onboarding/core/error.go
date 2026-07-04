package core

import "github.com/enviniom/nexokit/internal/platform/apperror"

// Module-owned business codes for the onboarding module.
const (
	CodeDuplicateCompanySlug     apperror.Code = "code:duplicate_company_slug"
	CodeDuplicateCompanyDomain   apperror.Code = "code:duplicate_company_domain"
	CodeDuplicateTechnicalDomain apperror.Code = "code:duplicate_technical_domain"
	CodeMissingPlatformDomain    apperror.Code = "code:missing_platform_domain"
	CodeDuplicateAdminEmail      apperror.Code = "code:duplicate_admin_email"
)

// Sentinel errors for the onboarding module.
//
// These are validation-style outcomes: the request as a whole is semantically
// invalid because a field (or field combination) conflicts with existing data
// or cannot be satisfied. Keeping them as apperror.Validation preserves the
// prior public HTTP contract of 422 Unprocessable Entity with field-keyed
// errors produced by the handler's thin mapping helper.
var (
	ErrDuplicateCompanySlug     = apperror.Validation(CodeDuplicateCompanySlug, "company slug already exists", nil)
	ErrDuplicateCompanyDomain   = apperror.Validation(CodeDuplicateCompanyDomain, "company domain already exists", nil)
	ErrDuplicateTechnicalDomain = apperror.Validation(CodeDuplicateTechnicalDomain, "company technical domain already exists", nil)
	ErrMissingPlatformDomain    = apperror.Validation(CodeMissingPlatformDomain, "platform domain is required to generate technical domain", nil)
	ErrDuplicateAdminEmail      = apperror.Validation(CodeDuplicateAdminEmail, "admin email already exists", nil)
)
