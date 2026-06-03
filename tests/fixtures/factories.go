package fixtures

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/modules/companies"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/enviniom/nexokit/internal/shared"
)

func BuildCompany(slug string) companies.Company {
	if slug == "" {
		slug = "acme"
	}
	return companies.Company{
		BaseModel: shared.BaseModel{PublicID: fmt.Sprintf("company-%s", slug)},
		Name:      fmt.Sprintf("Company %s", slug),
		Slug:      slug,
		Status:    companies.CompanyStatusActive,
	}
}

func BuildRole(slug string) iamcore.IAMRole {
	if slug == "" {
		slug = iamcore.UserRoleSlug
	}
	return iamcore.IAMRole{
		BaseModel:   shared.BaseModel{PublicID: fmt.Sprintf("role-%s", slug)},
		Name:        fmt.Sprintf("Role %s", slug),
		Slug:        slug,
		Description: fmt.Sprintf("Role for %s", slug),
		IsSystem:    true,
	}
}

func BuildPermission(module, action string) iamcore.IAMPermission {
	if module == "" {
		module = "users"
	}
	if action == "" {
		action = platformPerms.ActionView
	}
	slug := platformPerms.Format(module, action)
	return iamcore.IAMPermission{
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

func BuildUser(publicID string, roleID uint, companyID *uint) iamcore.IAMUser {
	if publicID == "" {
		publicID = "user-fixture"
	}
	return iamcore.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Name:         publicID,
		Email:        fmt.Sprintf("%s@example.com", publicID),
		PasswordHash: "hash",
		RoleID:       roleID,
		CompanyID:    companyID,
		IsActive:     true,
	}
}
