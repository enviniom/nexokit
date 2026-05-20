// Package tenant provides request-scoped tenant context and GORM helpers.
//
// Multitenant models should include a company_id column and every repository
// query for tenant-owned data should call ApplyTenantScope before executing the
// query. Root requests use NewRoot and remain unfiltered; non-root requests use
// NewScoped so data is constrained to one company.
package tenant
