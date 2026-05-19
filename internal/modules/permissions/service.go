package permissions

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// Service defines the business logic contract for permissions.
type Service interface {
	ListGrouped() ([]PermissionGroupResponse, error)
	List(page, perPage int) ([]PermissionResponse, int64, error)
	GetByPublicID(publicID string) (*PermissionResponse, error)
	Create(req CreatePermissionRequest) (*PermissionResponse, error)
	Update(publicID string, req UpdatePermissionRequest) (*PermissionResponse, error)
	Delete(publicID string) error
}

// permissionService is the concrete implementation of Service.
type permissionService struct {
	repo Repository
}

// NewService creates a new permissions service.
func NewService(repo Repository) Service {
	return &permissionService{repo: repo}
}

// ListGrouped returns all permissions grouped by module and sorted for rendering.
func (s *permissionService) ListGrouped() ([]PermissionGroupResponse, error) {
	items, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	return groupPermissions(items), nil
}

// List returns paginated permissions as DTOs.
func (s *permissionService) List(page, perPage int) ([]PermissionResponse, int64, error) {
	items, err := s.repo.List(page, perPage)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count()
	if err != nil {
		return nil, 0, err
	}
	responses := make([]PermissionResponse, len(items))
	for i := range items {
		responses[i] = *toResponse(&items[i])
	}
	return responses, total, nil
}

// GetByPublicID returns a single permission by public ID.
func (s *permissionService) GetByPublicID(publicID string) (*PermissionResponse, error) {
	permission, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return toResponse(permission), nil
}

// Create creates a non-system permission.
func (s *permissionService) Create(req CreatePermissionRequest) (*PermissionResponse, error) {
	if err := validatePermissionParts(req.Module, req.Action, req.Slug); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetBySlug(req.Slug); err == nil {
		return nil, apperror.ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	permission := &Permission{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Slug:         req.Slug,
		Name:         req.Name,
		Module:       req.Module,
		Action:       req.Action,
		Description:  req.Description,
		IsSystem:     false,
		DisplayOrder: req.DisplayOrder,
	}
	if err := s.repo.Create(permission); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}
	return toResponse(permission), nil
}

// Update updates a non-system permission.
func (s *permissionService) Update(publicID string, req UpdatePermissionRequest) (*PermissionResponse, error) {
	permission, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	if permission.IsSystem {
		return nil, apperror.ErrForbidden
	}
	if err := validatePermissionParts(req.Module, req.Action, req.Slug); err != nil {
		return nil, err
	}
	if permission.Slug != req.Slug {
		existing, err := s.repo.GetBySlug(req.Slug)
		if err == nil {
			if existing.PublicID != publicID {
				return nil, apperror.ErrConflict
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	permission.Slug = req.Slug
	permission.Name = req.Name
	permission.Module = req.Module
	permission.Action = req.Action
	permission.Description = req.Description
	permission.DisplayOrder = req.DisplayOrder

	if err := s.repo.Update(permission); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}
	return toResponse(permission), nil
}

// Delete deletes a non-system permission.
func (s *permissionService) Delete(publicID string) error {
	permission, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	if permission.IsSystem {
		return apperror.ErrForbidden
	}
	return s.repo.Delete(publicID)
}

func validatePermissionParts(module, action, slug string) error {
	if action == "read" {
		return apperror.ErrBadRequest
	}
	if module == "" || action == "" || slug == "" {
		return apperror.ErrBadRequest
	}
	if slug != fmt.Sprintf("%s.%s", module, action) {
		return apperror.ErrBadRequest
	}
	return nil
}

func groupPermissions(items []Permission) []PermissionGroupResponse {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Module != items[j].Module {
			return items[i].Module < items[j].Module
		}
		if items[i].DisplayOrder != items[j].DisplayOrder {
			return items[i].DisplayOrder < items[j].DisplayOrder
		}
		return items[i].Slug < items[j].Slug
	})

	groups := make([]PermissionGroupResponse, 0)
	moduleIndex := make(map[string]int)
	for i := range items {
		idx, ok := moduleIndex[items[i].Module]
		if !ok {
			groups = append(groups, PermissionGroupResponse{Module: items[i].Module})
			idx = len(groups) - 1
			moduleIndex[items[i].Module] = idx
		}
		groups[idx].Permissions = append(groups[idx].Permissions, *toResponse(&items[i]))
	}
	return groups
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") || strings.Contains(message, "unique failed")
}

func toResponse(p *Permission) *PermissionResponse {
	return &PermissionResponse{
		PublicID:     p.PublicID,
		Slug:         p.Slug,
		Name:         p.Name,
		Module:       p.Module,
		Action:       p.Action,
		Description:  p.Description,
		IsSystem:     p.IsSystem,
		DisplayOrder: p.DisplayOrder,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		CreatedBy:    p.CreatedBy,
		UpdatedBy:    p.UpdatedBy,
	}
}
