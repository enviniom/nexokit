package roles

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// Service defines the business logic contract for roles.
type Service interface {
	List(page, perPage int) ([]RoleResponse, int64, error)
	GetByPublicID(publicID string) (*RoleResponse, error)
	Create(req CreateRoleRequest) (*RoleResponse, error)
	Update(publicID string, req UpdateRoleRequest) (*RoleResponse, error)
	Delete(publicID string) error
}

// roleService is the concrete implementation of Service.
type roleService struct {
	repo Repository
}

// NewService creates a new roles service.
func NewService(repo Repository) Service {
	return &roleService{repo: repo}
}

// List returns paginated roles as DTOs.
func (s *roleService) List(page, perPage int) ([]RoleResponse, int64, error) {
	roles, err := s.repo.List(page, perPage)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count()
	if err != nil {
		return nil, 0, err
	}

	result := make([]RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = *toResponse(&r)
	}

	return result, total, nil
}

// GetByPublicID returns a single role by public ID.
func (s *roleService) GetByPublicID(publicID string) (*RoleResponse, error) {
	role, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return toResponse(role), nil
}

// Create creates a new role after checking name uniqueness.
func (s *roleService) Create(req CreateRoleRequest) (*RoleResponse, error) {
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, apperror.ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	role := &Role{
		BaseModel: shared.BaseModel{
			PublicID: publicID,
		},
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := s.repo.Create(role); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	return toResponse(role), nil
}

// Update updates a role if it is not a system role and the new name is unique.
func (s *roleService) Update(publicID string, req UpdateRoleRequest) (*RoleResponse, error) {
	role, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	if role.IsSystem {
		return nil, apperror.ErrForbidden
	}

	if role.Name != req.Name {
		existing, err := s.repo.GetByName(req.Name)
		if err == nil {
			if existing.PublicID != publicID {
				return nil, apperror.ErrConflict
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	role.Name = req.Name
	role.Slug = req.Slug
	role.Description = req.Description

	if err := s.repo.Update(role); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	return toResponse(role), nil
}

// Delete deletes a role if it is not a system role.
func (s *roleService) Delete(publicID string) error {
	role, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	if role.IsSystem {
		return apperror.ErrForbidden
	}

	return s.repo.Delete(publicID)
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

func toResponse(r *Role) *RoleResponse {
	return &RoleResponse{
		PublicID:    r.PublicID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CreatedBy:   r.CreatedBy,
		UpdatedBy:   r.UpdatedBy,
	}
}
