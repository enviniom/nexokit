package update_permission

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

// Service defines the business logic contract for permissions.
type Service interface {
	Update(publicID string, req core.UpdatePermissionRequest) (*core.PermissionResponse, error)
}

// permissionService is the concrete implementation of Service.
type permissionService struct {
	repo Repository
}

// NewService creates a new permissions service.
func NewService(repo Repository) Service {
	s := &permissionService{repo: repo}
	return s
}

// Update updates a non-system permission.
func (s *permissionService) Update(publicID string, req core.UpdatePermissionRequest) (*core.PermissionResponse, error) {
	permission, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	if permission.IsSystem {
		return nil, core.ErrSystemImmutable
	}

	permission.Name = req.Name
	permission.Description = req.Description
	permission.DisplayOrder = req.DisplayOrder

	if err := s.repo.Update(permission); err != nil {
		if isUniqueConstraintError(err) {
			return nil, core.ErrConflict
		}
		return nil, err
	}
	return core.ToResponse(permission), nil
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") || strings.Contains(message, "unique failed")
}
