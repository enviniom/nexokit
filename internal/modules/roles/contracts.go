package roles

import "github.com/enviniom/nexokit/internal/platform/tenant"

// AssignmentRoleSummary is a lightweight DTO containing metadata needed by
// other modules for role validation and assignment.
type AssignmentRoleSummary struct {
	InternalID uint
	PublicID   string
	Slug       string
	CompanyID  *uint
}

// AssignmentRoleReader defines the contract that other modules can consume
// to look up assignable roles within a tenant context.
type AssignmentRoleReader interface {
	FindAssignableByPublicID(tc tenant.TenantContext, publicID string) (*AssignmentRoleSummary, error)
}
