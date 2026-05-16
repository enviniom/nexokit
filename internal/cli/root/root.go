package root

import (
	"errors"
	"fmt"
	"net/mail"
	"unicode"
)

// ErrRootAlreadyExists is returned when a root user already exists.
var ErrRootAlreadyExists = errors.New("root user already exists")

// ErrStorageNotWired is returned when the root creator has no storage or hasher.
var ErrStorageNotWired = errors.New("root storage or password hasher not wired")

// CreateRootInput holds validated root creation parameters.
type CreateRootInput struct {
	Name     string
	Email    string
	Password string
}

// RootStorage defines the storage boundary for root creation.
type RootStorage interface {
	// RootExists returns true if any root user already exists.
	RootExists() (bool, error)
	// CreateRoot persists a new root user within a transaction.
	CreateRoot(name, email, passwordHash string) error
}

// PasswordHasher defines the boundary for password hashing.
// The concrete implementation will be provided by the password hashing change.
type PasswordHasher interface {
	// Hash returns a secure hash of the given password.
	Hash(password string) (string, error)
}

// Creator validates input and delegates to storage.
type Creator struct {
	Storage RootStorage
	Hasher  PasswordHasher
}

// NewCreator creates a root creator with the given storage and hasher.
// Both must be non-nil for Create to succeed; otherwise it returns ErrStorageNotWired.
func NewCreator(storage RootStorage, hasher PasswordHasher) *Creator {
	return &Creator{Storage: storage, Hasher: hasher}
}

// Create validates input and attempts to create a root user.
// It is idempotent: if a root already exists, it returns ErrRootAlreadyExists.
// If storage or hasher is not wired, it returns ErrStorageNotWired.
func (c *Creator) Create(input CreateRootInput) error {
	if err := ValidateInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if c.Storage == nil || c.Hasher == nil {
		return ErrStorageNotWired
	}

	exists, err := c.Storage.RootExists()
	if err != nil {
		return fmt.Errorf("failed to check root existence: %w", err)
	}
	if exists {
		return ErrRootAlreadyExists
	}

	hash, err := c.Hasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := c.Storage.CreateRoot(input.Name, input.Email, hash); err != nil {
		return fmt.Errorf("failed to create root user: %w", err)
	}

	return nil
}

// ValidateInput checks name, email and password requirements.
func ValidateInput(input CreateRootInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if input.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}
	if input.Password == "" {
		return errors.New("password is required")
	}
	if len(input.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !hasMixedCaseAndDigit(input.Password) {
		return errors.New("password must contain uppercase, lowercase, and a digit")
	}
	return nil
}

func hasMixedCaseAndDigit(s string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}
