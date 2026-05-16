package users

import (
	"errors"
	"strings"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// PasswordHasher defines the boundary for password hashing.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// Service defines the business logic contract for users.
type Service interface {
	List(page, perPage int) ([]UserResponse, int64, error)
	GetByPublicID(publicID string) (*UserResponse, error)
	Create(req CreateUserRequest) (*UserResponse, error)
	Update(publicID string, req UpdateUserRequest) (*UserResponse, error)
	Delete(publicID string) error
	ChangePassword(publicID string, req ChangePasswordRequest) error
	ToggleStatus(publicID string, req UpdateStatusRequest) (*UserResponse, error)
}

// userService is the concrete implementation of Service.
type userService struct {
	repo   Repository
	hasher PasswordHasher
}

// NewService creates a new users service.
func NewService(repo Repository, hasher PasswordHasher) Service {
	return &userService{repo: repo, hasher: hasher}
}

// List returns paginated users as DTOs.
func (s *userService) List(page, perPage int) ([]UserResponse, int64, error) {
	users, err := s.repo.List(page, perPage)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count()
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
func (s *userService) GetByPublicID(publicID string) (*UserResponse, error) {
	user, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return toResponse(user), nil
}

// Create creates a new user after checking email uniqueness.
func (s *userService) Create(req CreateUserRequest) (*UserResponse, error) {
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, apperror.ErrConflict
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(req.Password)
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

	created, err := s.repo.GetByPublicID(user.PublicID)
	if err != nil {
		return nil, err
	}

	return toResponse(created), nil
}

// Update updates a user if the new email is unique.
func (s *userService) Update(publicID string, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
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
	user.RoleID = req.RoleID
	user.CompanyID = req.CompanyID

	if err := s.repo.Update(user); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	updated, err := s.repo.GetByPublicID(user.PublicID)
	if err != nil {
		return nil, err
	}

	return toResponse(updated), nil
}

// Delete soft-deletes a user by its public ID.
func (s *userService) Delete(publicID string) error {
	user, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	_ = user // reserved for future audit logic
	return s.repo.Delete(publicID)
}

// ChangePassword verifies the current password and updates the hash.
func (s *userService) ChangePassword(publicID string, req ChangePasswordRequest) error {
	user, err := s.repo.GetByPublicID(publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	if err := s.hasher.Verify(req.CurrentPassword, user.PasswordHash); err != nil {
		return apperror.ErrUnauthorized
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	return s.repo.Update(user)
}

// ToggleStatus updates a user's active status.
func (s *userService) ToggleStatus(publicID string, req UpdateStatusRequest) (*UserResponse, error) {
	user, err := s.repo.GetByPublicID(publicID)
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

	updated, err := s.repo.GetByPublicID(user.PublicID)
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
