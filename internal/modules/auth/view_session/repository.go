package view_session

import "github.com/enviniom/nexokit/internal/platform/authctx"

type SessionView struct {
	PublicID    string
	Name        string
	Email       string
	IsActive    bool
	RoleID      uint
	RoleName    string
	RoleSlug    string
	CompanyID   *uint
	Permissions []string
}

type Repository interface {
	BuildSession(current *authctx.User) (*SessionView, error)
}

type repository struct{}

func NewRepository() Repository { return repository{} }

func (repository) BuildSession(current *authctx.User) (*SessionView, error) {
	return &SessionView{
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
