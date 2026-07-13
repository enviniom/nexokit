package view_session

import "github.com/enviniom/nexokit/internal/platform/authctx"

type Service interface {
	View(current *authctx.User) (*SessionView, error)
}

type service struct{ repository Repository }

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) View(current *authctx.User) (*SessionView, error) {
	return s.repository.BuildSession(current)
}
