package change_user_password

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

// PasswordHasher abstracts password hashing for the change_user_password slice.
type PasswordHasher interface {
	HashPassword(string) (string, error)
	VerifyPassword(password, hash string) error
}

// Service orchestrates the change_user_password use case.
type Service interface {
	Change(tc tenant.TenantContext, publicID, actorPublicID string, req core.ChangePasswordRequest) error
}

type service struct {
	repo   Repository
	hasher PasswordHasher
}

// NewService creates a change_user_password service backed by the given repository and hasher.
func NewService(repo Repository, hasher PasswordHasher) Service {
	return &service{repo: repo, hasher: hasher}
}

func (s *service) Change(tc tenant.TenantContext, publicID, actorPublicID string, req core.ChangePasswordRequest) error {
	u, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return err
	}

	rootRole, err := s.repo.GetRoleBySlug(core.RootRoleSlug)
	if err != nil {
		return err
	}

	if u.RoleID == rootRole.ID && (actorPublicID == "" || actorPublicID != u.PublicID) {
		return core.ErrForbidden
	}

	if err := s.hasher.VerifyPassword(req.CurrentPassword, u.PasswordHash); err != nil {
		return core.ErrUnauthorized
	}

	hash, err := s.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(u.ID, hash)
}
