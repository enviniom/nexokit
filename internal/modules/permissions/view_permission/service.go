package view_permission

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"gorm.io/gorm"
)

// Service defines the business logic contract for permissions.
type Service interface {
	GetByPublicID(publicID string) (*core.PermissionResponse, error)
}

// permissionService is the concrete implementation of Service.
type permissionService struct {
	repo Repository
}

// ServiceOption configures optional permission service collaborators.
type ServiceOption func(*permissionService)

// NewService creates a new permissions service.
func NewService(repo Repository) Service {
	s := &permissionService{repo: repo}
	return s
}

// GetByPublicID returns a single permission by public ID.
func (s *permissionService) GetByPublicID(publicID string) (*core.PermissionResponse, error) {
	permission, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return core.ToResponse(permission), nil
}
