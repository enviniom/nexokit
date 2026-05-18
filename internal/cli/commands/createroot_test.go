package commands

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

// fakeRootStorage is a test double for root.RootStorage.
type fakeRootStorage struct {
	exists    bool
	existsErr error
	createErr error
	created   bool
	lastName  string
	lastEmail string
	lastHash  string
}

func (f *fakeRootStorage) RootExists() (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeRootStorage) CreateRoot(name, email, passwordHash string) error {
	f.created = true
	f.lastName = name
	f.lastEmail = email
	f.lastHash = passwordHash
	return f.createErr
}

// fakeRootHasher is a test double for root.PasswordHasher.
type fakeRootHasher struct {
	hash string
	err  error
}

func (f *fakeRootHasher) HashPassword(password string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.hash + password, nil
}

func TestCreateRootCommand_WithFlags(t *testing.T) {
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var out bytes.Buffer
	stdio := cli.Stdio{In: strings.NewReader(""), Out: &out, Err: &out}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1"}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.created {
		t.Error("expected storage.CreateRoot to be called")
	}
	if storage.lastName != "Root" {
		t.Errorf("expected name 'Root', got %q", storage.lastName)
	}
	if storage.lastEmail != "root@test.com" {
		t.Errorf("expected email 'root@test.com', got %q", storage.lastEmail)
	}
	if storage.lastHash != "mock-hash:Password1" {
		t.Errorf("expected hashed password, got %q", storage.lastHash)
	}
	if !strings.Contains(out.String(), "created successfully") {
		t.Errorf("expected success message, got %q", out.String())
	}
}

func TestCreateRootCommand_Prompts(t *testing.T) {
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	in := strings.NewReader("Root\nroot@test.com\nPassword1\n")
	var out bytes.Buffer
	stdio := cli.Stdio{In: in, Out: &out}
	err := cmd.Run(context.Background(), []string{}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.created {
		t.Error("expected storage.CreateRoot to be called")
	}
}

func TestCreateRootCommand_InvalidEmail(t *testing.T) {
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "bad", "--password", "Password1"}, stdio)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid email") {
		t.Errorf("expected invalid email error, got %v", err)
	}
}

func TestCreateRootCommand_WeakPassword(t *testing.T) {
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "short"}, stdio)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("expected password error, got %v", err)
	}
}

func TestCreateRootCommand_ConfirmationAborted(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	in := strings.NewReader("Root\nroot@test.com\nPassword1\nno\n")
	var out bytes.Buffer
	stdio := cli.Stdio{In: in, Out: &out}
	err := cmd.Run(context.Background(), []string{}, stdio)
	if err == nil {
		t.Fatal("expected aborted error")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("expected aborted error, got %v", err)
	}
}

func TestCreateRootCommand_ForceSkipConfirm(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var out bytes.Buffer
	var errOut bytes.Buffer
	stdio := cli.Stdio{Out: &out, Err: &errOut}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1", "--force"}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(errOut.String(), "aborted") {
		t.Error("should not abort when --force is set")
	}
}

func TestCreateRootCommand_EnvVars(t *testing.T) {
	t.Setenv("ROOT_USER_NAME", "EnvRoot")
	t.Setenv("ROOT_USER_EMAIL", "env@example.com")
	t.Setenv("ROOT_USER_PASSWORD", "Password1")

	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var out bytes.Buffer
	stdio := cli.Stdio{In: strings.NewReader(""), Out: &out}
	err := cmd.Run(context.Background(), []string{}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.created {
		t.Error("expected storage.CreateRoot to be called")
	}
	if storage.lastName != "EnvRoot" {
		t.Errorf("expected name 'EnvRoot', got %q", storage.lastName)
	}
	if storage.lastEmail != "env@example.com" {
		t.Errorf("expected email 'env@example.com', got %q", storage.lastEmail)
	}
}

func TestCreateRootCommand_Idempotent(t *testing.T) {
	storage := &fakeRootStorage{exists: true}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var out bytes.Buffer
	stdio := cli.Stdio{In: strings.NewReader(""), Out: &out}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1"}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("expected idempotency message, got %q", out.String())
	}
	if storage.created {
		t.Error("storage.CreateRoot should not be called when root already exists")
	}
}

func TestCreateRootCommand_StorageError(t *testing.T) {
	storage := &fakeRootStorage{createErr: errors.New("db down")}
	hasher := &fakeRootHasher{hash: "mock-hash:"}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1"}, stdio)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "db down") {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestCreateRootCommand_HasherError(t *testing.T) {
	storage := &fakeRootStorage{}
	hasher := &fakeRootHasher{err: errors.New("hash error")}
	cmd := &CreateRootCommand{Storage: storage, Hasher: hasher}

	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	err := cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1"}, stdio)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "hash error") {
		t.Errorf("expected hash error, got %v", err)
	}
}
