package core

import (
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
)

// Module-owned business codes for the companies module.
const (
	CodeCompanyNotFound            apperror.Code = "code:company_not_found"
	CodeCompanyDomainNotFound      apperror.Code = "code:company_domain_not_found"
	CodeCompanyDomainDuplicate     apperror.Code = "code:company_domain_duplicate"
	CodePrimaryDomainExists        apperror.Code = "code:primary_domain_exists"
	CodeCompanyDomainDoesNotBelong apperror.Code = "code:company_domain_does_not_belong"
	CodeCompanySlugDuplicate       apperror.Code = "code:company_slug_duplicate"
	CodeCompanyPersistence         apperror.Code = "code:company_persistence_error"
	CodeCompanyDomainPersistence   apperror.Code = "code:company_domain_persistence_error"
)

// CompanyPersistenceError reports an unexpected company persistence failure
// while retaining its original cause for logging.
func CompanyPersistenceError(cause error) error {
	return apperror.Internal(CodeCompanyPersistence, messages.MsgInternalError, cause)
}

// CompanyDomainPersistenceError reports an unexpected company-domain
// persistence failure while retaining its original cause for logging.
func CompanyDomainPersistenceError(cause error) error {
	return apperror.Internal(CodeCompanyDomainPersistence, messages.MsgInternalError, cause)
}

// Sentinel errors for the companies module.
//
// Duplicate domain/slug conflicts are modeled as 409 Conflict so the sentinel
// status reflects the resource conflict semantics required by the change
// acceptance criteria. The create/update handlers preserve the original public
// contract by mapping those sentinels to field-keyed 422 validation responses.
// Active-primary-domain and ownership failures keep the status that matches
// their existing handler presentation.
var (
	ErrCompanyNotFound            = apperror.NotFound(CodeCompanyNotFound, "company not found", nil)
	ErrCompanyDomainNotFound      = apperror.NotFound(CodeCompanyDomainNotFound, "company domain not found", nil)
	ErrDuplicateCompanyDomain     = apperror.Conflict(CodeCompanyDomainDuplicate, "company domain already exists", nil)
	ErrActivePrimaryDomainExists  = apperror.Validation(CodePrimaryDomainExists, "company already has an active primary domain", nil)
	ErrCompanyDomainDoesNotBelong = apperror.NotFound(CodeCompanyDomainDoesNotBelong, "company domain does not belong to company", nil)
	ErrDuplicateCompanySlug       = apperror.Conflict(CodeCompanySlugDuplicate, "company slug already exists", nil)
)
