package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// PasswordHasher defines the boundary for password hashing.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
}

// RoleResolver resolves role metadata for business rule enforcement.
type RoleResolver interface {
	GetBySlug(slug string) (*roles.Role, error)
}

// Service defines the business logic contract for users.
type Service interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]UserResponse, int64, error)
	GetByPublicID(tc tenant.TenantContext, publicID string) (*UserResponse, error)
	Create(tc tenant.TenantContext, req CreateUserRequest) (*UserResponse, error)
	Update(tc tenant.TenantContext, publicID string, actorPublicID string, req UpdateUserRequest) (*UserResponse, error)
	Delete(tc tenant.TenantContext, publicID string) error
	ChangePassword(tc tenant.TenantContext, publicID string, actorPublicID string, req ChangePasswordRequest) error
	ToggleStatus(tc tenant.TenantContext, publicID string, req UpdateStatusRequest) (*UserResponse, error)
	ChangeRole(tc tenant.TenantContext, targetPublicID string, actorPublicID string, req ChangeUserRoleRequest) (*UserResponse, error)
}

// userService is the concrete implementation of Service.
type userService struct {
	repo             Repository
	hasher           PasswordHasher
	resolver         RoleResolver
	cachedRootRoleID uint
	roleReader       roles.AssignmentRoleReader
	cache            cache.Cache
}

// ServiceOption configures optional user service collaborators.
type ServiceOption func(*userService)

// WithRoleReader configures the role reader collaborator.
func WithRoleReader(r roles.AssignmentRoleReader) ServiceOption {
	return func(s *userService) {
		s.roleReader = r
	}
}

// WithCache configures the cache collaborator.
func WithCache(c cache.Cache) ServiceOption {
	return func(s *userService) {
		s.cache = c
	}
}

// NewService creates a new users service.
func NewService(repo Repository, hasher PasswordHasher, resolver RoleResolver, opts ...ServiceOption) Service {
	s := &userService{repo: repo, hasher: hasher, resolver: resolver}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *userService) rootRoleID() (uint, error) {
	if s.cachedRootRoleID != 0 {
		return s.cachedRootRoleID, nil
	}
	if s.resolver == nil {
		return 0, errors.New("role resolver not configured")
	}
	role, err := s.resolver.GetBySlug(roles.RootRoleSlug)
	if err != nil {
		return 0, err
	}
	if role == nil {
		return 0, errors.New("root role not found")
	}
	s.cachedRootRoleID = role.ID
	return role.ID, nil
}

// List returns paginated users as DTOs.
func (s *userService) List(tc tenant.TenantContext, params query.ListParams) ([]UserResponse, int64, error) {
	users, err := s.repo.List(tc, params)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(tc, params)
	if err != nil {
		return nil, 0, err
	}

	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = *toResponse(&u)
	}

	return result, total, nil
}

// GetByPublicID returns a single user by public ID.
func (s *userService) GetByPublicID(tc tenant.TenantContext, publicID string) (*UserResponse, error) {
	user, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return toResponse(user), nil
}

// Create creates a new user after checking email uniqueness.
func (s *userService) Create(tc tenant.TenantContext, req CreateUserRequest) (*UserResponse, error) {
	rootRoleID, err := s.rootRoleID()
	if err != nil {
		return nil, err
	}

	if req.RoleID == rootRoleID {
		if !tc.IsRootScope || req.CompanyID != nil {
			return nil, apperror.ErrForbidden
		}
	} else if tc.IsRootScope {
		if req.CompanyID == nil {
			return nil, apperror.ErrBadRequest
		}
	} else {
		if req.CompanyID != nil && *req.CompanyID != tc.CompanyID {
			return nil, apperror.ErrForbidden
		}
		companyID := tc.CompanyID
		req.CompanyID = &companyID
	}

	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, apperror.ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	hash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		BaseModel: shared.BaseModel{
			PublicID: publicID,
		},
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		RoleID:       req.RoleID,
		CompanyID:    req.CompanyID,
		IsActive:     true,
	}

	if err := s.repo.Create(user); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	created, err := s.repo.GetByPublicID(tenant.NewRoot(), user.PublicID)
	if err != nil {
		return nil, err
	}

	return toResponse(created), nil
}

// Update updates a user if the new email is unique.
func (s *userService) Update(tc tenant.TenantContext, publicID string, actorPublicID string, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	rootRoleID, err := s.rootRoleID()
	if err != nil {
		return nil, err
	}

	isRoot := user.RoleID == rootRoleID

	if isRoot {
		// Rule: root can only be edited by itself.
		// When auth context is unavailable (empty actor), reject for safety.
		if actorPublicID == "" || actorPublicID != user.PublicID {
			return nil, apperror.ErrForbidden
		}
		// Rule: root can only have name and email edited.
		user.Name = req.Name
		user.Email = req.Email
		// RoleID and CompanyID are intentionally ignored for root.
	} else {
		// Rule: non-root users in a tenant must not be moved out of their company
		if !tc.IsRootScope {
			if req.CompanyID != nil && *req.CompanyID != tc.CompanyID {
				return nil, apperror.ErrForbidden
			}
			companyID := tc.CompanyID
			req.CompanyID = &companyID
		} else {
			// In root scope, we must ensure CompanyID is not nil for a non-root user.
			if req.CompanyID == nil {
				return nil, apperror.ErrBadRequest
			}
		}

		if user.Email != req.Email {
			existing, err := s.repo.GetByEmail(req.Email)
			if err == nil {
				if existing.PublicID != publicID {
					return nil, apperror.ErrConflict
				}
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}

		user.Name = req.Name
		user.Email = req.Email
		user.CompanyID = req.CompanyID
	}

	if err := s.repo.Update(user); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	updated, err := s.repo.GetByPublicID(tc, user.PublicID)
	if err != nil {
		return nil, err
	}

	return toResponse(updated), nil
}

// Delete soft-deletes a user by its public ID.
func (s *userService) Delete(tc tenant.TenantContext, publicID string) error {
	user, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	_ = user // reserved for future audit logic
	if err := s.repo.Delete(tc, publicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	return nil
}

// ChangePassword verifies the current password and updates the hash.
func (s *userService) ChangePassword(tc tenant.TenantContext, publicID string, actorPublicID string, req ChangePasswordRequest) error {
	user, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	rootRoleID, err := s.rootRoleID()
	if err != nil {
		return err
	}

	// Rule: root can only change its own password.
	if user.RoleID == rootRoleID {
		if actorPublicID == "" || actorPublicID != user.PublicID {
			return apperror.ErrForbidden
		}
	}

	if err := s.hasher.VerifyPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		return apperror.ErrUnauthorized
	}

	newHash, err := s.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	return s.repo.Update(user)
}

// ToggleStatus updates a user's active status.
func (s *userService) ToggleStatus(tc tenant.TenantContext, publicID string, req UpdateStatusRequest) (*UserResponse, error) {
	user, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	user.IsActive = req.IsActive
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetByPublicID(tc, user.PublicID)
	if err != nil {
		return nil, err
	}

	return toResponse(updated), nil
}

// ChangeRole changes a user's role.
func (s *userService) ChangeRole(tc tenant.TenantContext, targetPublicID string, actorPublicID string, req ChangeUserRoleRequest) (*UserResponse, error) {
	if actorPublicID == "" || actorPublicID == targetPublicID {
		return nil, apperror.ErrForbidden
	}

	targetUser, err := s.repo.GetByPublicID(tc, targetPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	rootRoleID, err := s.rootRoleID()
	if err != nil {
		return nil, err
	}

	if targetUser.RoleID == rootRoleID {
		return nil, apperror.ErrForbidden
	}

	if s.roleReader == nil {
		return nil, errors.New("role reader dependency is not configured")
	}

	targetRole, err := s.roleReader.FindAssignableByPublicID(tc, req.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	if targetRole.Slug == roles.RootRoleSlug {
		return nil, apperror.ErrForbidden
	}

	if targetUser.CompanyID != nil && targetRole.CompanyID != nil && *targetUser.CompanyID != *targetRole.CompanyID {
		return nil, apperror.ErrForbidden
	}

	targetUser.RoleID = targetRole.InternalID
	if err := s.repo.Update(targetUser); err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Delete(context.Background(), fmt.Sprintf("rbac:permissions:%s", targetUser.PublicID))
	}

	updated, err := s.repo.GetByPublicID(tc, targetUser.PublicID)
	if err != nil {
		return nil, err
	}

	return toResponse(updated), nil
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

func toResponse(u *User) *UserResponse {
	return &UserResponse{
		PublicID:  u.PublicID,
		Name:      u.Name,
		Email:     u.Email,
		IsActive:  u.IsActive,
		RoleID:    u.RoleID,
		RoleName:  u.Role.Name,
		CompanyID: u.CompanyID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		CreatedBy: u.CreatedBy,
		UpdatedBy: u.UpdatedBy,
	}
}
