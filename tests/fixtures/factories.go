package fixtures

import (
	"fmt"
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"

	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/enviniom/nexokit/internal/shared"
)

func BuildCompany(slug string) company.Company {
	if slug == "" {
		slug = "acme"
	}
	return company.Company{
		BaseModel: shared.BaseModel{PublicID: fmt.Sprintf("company-%s", slug)},
		Name:      fmt.Sprintf("Company %s", slug),
		Slug:      slug,
		Status:    core.CompanyStatusActive,
	}
}

func BuildRole(slug string) iamrole.IAMRole {
	if slug == "" {
		slug = iamcore.UserRoleSlug
	}
	return iamrole.IAMRole{
		BaseModel:   shared.BaseModel{PublicID: fmt.Sprintf("role-%s", slug)},
		Name:        fmt.Sprintf("Role %s", slug),
		Slug:        slug,
		Description: fmt.Sprintf("Role for %s", slug),
		IsSystem:    true,
	}
}

func BuildPermission(module, action string) permission.IAMPermission {
	if module == "" {
		module = "users"
	}
	if action == "" {
		action = platformPerms.ActionView
	}
	slug := platformPerms.Format(module, action)
	return permission.IAMPermission{
		BaseModel:    shared.BaseModel{PublicID: fmt.Sprintf("perm-%s-%s", module, action)},
		Slug:         slug,
		Name:         fmt.Sprintf("%s %s", module, action),
		Module:       module,
		Action:       action,
		Description:  fmt.Sprintf("Allows %s on %s", action, module),
		IsSystem:     true,
		DisplayOrder: 1,
	}
}

func BuildUser(publicID string, roleID uint, companyID *uint) iamuser.IAMUser {
	if publicID == "" {
		publicID = "user-fixture"
	}
	return iamuser.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Name:         publicID,
		Email:        fmt.Sprintf("%s@example.com", publicID),
		PasswordHash: "hash",
		RoleID:       roleID,
		CompanyID:    companyID,
		IsActive:     true,
	}
}
