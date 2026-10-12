package view_session

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"github.com/enviniom/nexokit/internal/platform/authctx"
)

type Repository interface {
	BuildSession(current *authctx.User) (*user.SessionView, error)
}

type repository struct{}

func NewRepository() Repository { return repository{} }

func (repository) BuildSession(current *authctx.User) (*user.SessionView, error) {
	return &user.SessionView{
		PublicID:    current.PublicID,
		Name:        current.Name,
		Email:       current.Email,
		IsActive:    current.IsActive,
		RoleID:      current.RoleID,
		RoleName:    current.Role,
		RoleSlug:    current.RoleSlug,
		CompanyID:   current.CompanyID,
		Permissions: current.Permissions,
	}, nil
}
