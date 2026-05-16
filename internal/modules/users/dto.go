package users

import "time"

// UserResponse is the DTO for user read operations.
type UserResponse struct {
	PublicID  string    `json:"public_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	RoleID    uint      `json:"role_id"`
	RoleName  string    `json:"role_name"`
	CompanyID *uint     `json:"company_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy *uint     `json:"created_by,omitempty"`
	UpdatedBy *uint     `json:"updated_by,omitempty"`
}

// CreateUserRequest is the DTO for creating a user.
type CreateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	RoleID    uint   `json:"role_id" binding:"required"`
	CompanyID *uint  `json:"company_id,omitempty"`
}

// UpdateUserRequest is the DTO for updating a user.
type UpdateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	RoleID    uint   `json:"role_id" binding:"required"`
	CompanyID *uint  `json:"company_id,omitempty"`
}

// ChangePasswordRequest is the DTO for changing a user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UpdateStatusRequest is the DTO for toggling a user's active status.
type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}
