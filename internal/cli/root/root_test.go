package root

import (
	"errors"
	"strings"
	"testing"
)

type mockStorage struct {
	exists         bool
	existsErr      error
	createErr      error
	created        bool
	lastEmail      string
	lastPasswordHash string
}

func (m *mockStorage) RootExists() (bool, error) {
	return m.exists, m.existsErr
}

func (m *mockStorage) CreateRoot(email, passwordHash string) error {
	m.created = true
	m.lastEmail = email
	m.lastPasswordHash = passwordHash
	return m.createErr
}

type mockHasher struct{}

func (m *mockHasher) Hash(password string) (string, error) {
	return "mock-hash:" + password, nil
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name  string
		input CreateRootInput
		want  string
	}{
		{"empty email", CreateRootInput{Password: "Password1"}, "email is required"},
		{"invalid email", CreateRootInput{Email: "not-an-email", Password: "Password1"}, "invalid email"},
		{"empty password", CreateRootInput{Email: "root@example.com"}, "password is required"},
		{"short password", CreateRootInput{Email: "root@example.com", Password: "Short1"}, "password must be at least 8 characters"},
		{"no uppercase", CreateRootInput{Email: "root@example.com", Password: "password1"}, "password must contain uppercase, lowercase, and a digit"},
		{"no lowercase", CreateRootInput{Email: "root@example.com", Password: "PASSWORD1"}, "password must contain uppercase, lowercase, and a digit"},
		{"no digit", CreateRootInput{Email: "root@example.com", Password: "Password"}, "password must contain uppercase, lowercase, and a digit"},
		{"valid", CreateRootInput{Email: "root@example.com", Password: "Password1"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInput(tt.input)
			if tt.want == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("expected error %q, got nil", tt.want)
				return
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestCreator_StorageNotWired(t *testing.T) {
	c := NewCreator(nil, nil)
	err := c.Create(CreateRootInput{Email: "root@example.com", Password: "Password1"})
	if !errors.Is(err, ErrStorageNotWired) {
		t.Errorf("expected ErrStorageNotWired, got %v", err)
	}
}

func TestCreator_StorageWithoutHasher(t *testing.T) {
	storage := &mockStorage{}
	c := NewCreator(storage, nil)
	err := c.Create(CreateRootInput{Email: "root@example.com", Password: "Password1"})
	if !errors.Is(err, ErrStorageNotWired) {
		t.Errorf("expected ErrStorageNotWired, got %v", err)
	}
	if storage.created {
		t.Error("storage.CreateRoot should not be called when hasher is nil")
	}
}

func TestCreator_Idempotent(t *testing.T) {
	storage := &mockStorage{exists: true}
	c := NewCreator(storage, &mockHasher{})
	err := c.Create(CreateRootInput{Email: "root@example.com", Password: "Password1"})
	if !errors.Is(err, ErrRootAlreadyExists) {
		t.Errorf("expected ErrRootAlreadyExists, got %v", err)
	}
}

func TestCreator_CreateSuccess(t *testing.T) {
	storage := &mockStorage{}
	c := NewCreator(storage, &mockHasher{})
	err := c.Create(CreateRootInput{Email: "root@example.com", Password: "Password1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.created {
		t.Error("expected storage.CreateRoot to be called")
	}
	if storage.lastPasswordHash == "Password1" {
		t.Error("expected hashed password, got raw password")
	}
	if storage.lastPasswordHash != "mock-hash:Password1" {
		t.Errorf("expected mock hash, got %q", storage.lastPasswordHash)
	}
}

func TestCreator_CreateStorageError(t *testing.T) {
	storage := &mockStorage{createErr: errors.New("db down")}
	c := NewCreator(storage, &mockHasher{})
	err := c.Create(CreateRootInput{Email: "root@example.com", Password: "Password1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreator_ExistsCheckError(t *testing.T) {
	storage := &mockStorage{existsErr: errors.New("timeout")}
	c := NewCreator(storage, &mockHasher{})
	err := c.Create(CreateRootInput{Email: "root@example.com", Password: "Password1"})
	if err == nil {
		t.Fatal("expected error")
	}
}
