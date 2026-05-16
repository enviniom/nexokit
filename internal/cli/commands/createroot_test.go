package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

func TestCreateRootCommand_WithFlags(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	stdio := cli.Stdio{In: strings.NewReader(""), Out: &out, Err: &errOut}
	cmd := &CreateRootCommand{}

	err := cmd.Run(context.Background(), []string{"--email", "root@test.com", "--password", "Password1"}, stdio)
	// Storage is not wired, so it returns an error
	if err == nil {
		t.Fatal("expected error because storage is not wired")
	}
	if !strings.Contains(err.Error(), "not yet wired") {
		t.Errorf("expected 'not yet wired' error, got %v", err)
	}
}

func TestCreateRootCommand_Prompts(t *testing.T) {
	in := strings.NewReader("root@test.com\nPassword1\n")
	var out bytes.Buffer
	stdio := cli.Stdio{In: in, Out: &out}
	cmd := &CreateRootCommand{}

	err := cmd.Run(context.Background(), []string{}, stdio)
	if err == nil {
		t.Fatal("expected error because storage is not wired")
	}
}

func TestCreateRootCommand_InvalidEmail(t *testing.T) {
	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	cmd := &CreateRootCommand{}

	err := cmd.Run(context.Background(), []string{"--email", "bad", "--password", "Password1"}, stdio)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid email") {
		t.Errorf("expected invalid email error, got %v", err)
	}
}

func TestCreateRootCommand_WeakPassword(t *testing.T) {
	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	cmd := &CreateRootCommand{}

	err := cmd.Run(context.Background(), []string{"--email", "root@test.com", "--password", "short"}, stdio)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("expected password error, got %v", err)
	}
}

func TestCreateRootCommand_ConfirmationAborted(t *testing.T) {
	// Simulate non-local environment by setting APP_ENV to production
	t.Setenv("APP_ENV", "production")
	in := strings.NewReader("root@test.com\nPassword1\nno\n")
	var out bytes.Buffer
	stdio := cli.Stdio{In: in, Out: &out}
	cmd := &CreateRootCommand{}

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
	var errOut bytes.Buffer
	stdio := cli.Stdio{Err: &errOut}
	cmd := &CreateRootCommand{}

	err := cmd.Run(context.Background(), []string{"--email", "root@test.com", "--password", "Password1", "--force"}, stdio)
	// Should reach storage-not-wired, not confirmation prompt
	if err == nil {
		t.Fatal("expected error because storage is not wired")
	}
	if strings.Contains(err.Error(), "aborted") {
		t.Error("should not abort when --force is set")
	}
}
