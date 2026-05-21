package fixtures

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
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

func BuildRole(slug string) roles.Role {
	if slug == "" {
		slug = roles.UserRoleSlug
	}
	return roles.Role{
		BaseModel:   shared.BaseModel{PublicID: fmt.Sprintf("role-%s", slug)},
		Name:        fmt.Sprintf("Role %s", slug),
		Slug:        slug,
		Description: fmt.Sprintf("Role for %s", slug),
		IsSystem:    true,
	}
}

func BuildPermission(module, action string) permissions.Permission {
	if module == "" {
		module = "users"
	}
	if action == "" {
		action = permissions.ActionView
	}
	slug := fmt.Sprintf("%s:%s", module, action)
	return permissions.Permission{
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

func BuildUser(publicID string, roleID uint, companyID *uint) users.User {
	if publicID == "" {
		publicID = "user-fixture"
	}
	return users.User{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Name:         publicID,
		Email:        fmt.Sprintf("%s@example.com", publicID),
		PasswordHash: "hash",
		RoleID:       roleID,
		CompanyID:    companyID,
		IsActive:     true,
	}
}
