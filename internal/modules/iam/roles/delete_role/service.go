package delete_role

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type Service interface {
	Delete(tc tenant.TenantContext, publicID string) error
}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }
func (s *service) Delete(tc tenant.TenantContext, publicID string) error {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		return err
	}
	if role.IsSystem || core.IsReservedRoleIdentity(role.Name, role.Slug) {
		return core.ErrRoleProtected
	}
	count, err := s.repo.CountUsersByRoleID(role.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return core.ErrRoleHasAssignedUsers
	}
	return s.repo.DeleteByPublicID(tc, publicID)
}
