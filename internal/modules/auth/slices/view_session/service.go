package view_session

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core/user"
	"github.com/enviniom/nexokit/internal/platform/authctx"
)

type Service interface {
	View(current *authctx.User) (*user.SessionView, error)
}

type service struct{ repository Repository }

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) View(current *authctx.User) (*user.SessionView, error) {
	return s.repository.BuildSession(current)
}
