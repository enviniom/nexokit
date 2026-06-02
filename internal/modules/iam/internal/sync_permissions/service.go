package sync_permissions

import (
	"strings"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/identity"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/enviniom/nexokit/internal/shared"
)

type Service interface{ SyncPermissions(slugs []string) error }

type Repository interface {
	FindBySlug(slug string) (*core.IAMPermission, bool, error)
	Create(permission *core.IAMPermission) error
	AutoAssignToAdmins(permissionID uint) error
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) SyncPermissions(slugs []string) error {
	for _, slug := range slugs {
		parts := strings.SplitN(slug, ".", 2)
		if len(parts) != 2 {
			continue
		}
		_, found, err := s.repo.FindBySlug(slug)
		if err != nil {
			return err
		}
		if found {
			continue
		}
		publicID, err := identity.Generate()
		if err != nil {
			return err
		}
		module, action := parts[0], parts[1]
		permission := &core.IAMPermission{
			BaseModel:    shared.BaseModel{PublicID: publicID},
			Slug:         slug,
			Name:         platformPerms.HumanizeName(module, action),
			Module:       module,
			Action:       action,
			Description:  platformPerms.HumanizeDescription(module, action),
			IsSystem:     true,
			DisplayOrder: platformPerms.DefaultDisplayOrder(action),
		}
		if err := s.repo.Create(permission); err != nil {
			return err
		}
		if err := s.repo.AutoAssignToAdmins(permission.ID); err != nil {
			return err
		}
	}
	return nil
}
