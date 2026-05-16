package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type testCommand struct {
	name    string
	desc    string
	runFunc func(context.Context, []string, Stdio) error
}

func (t *testCommand) Name() string        { return t.name }
func (t *testCommand) Description() string { return t.desc }
func (t *testCommand) Run(ctx context.Context, args []string, stdio Stdio) error {
	if t.runFunc != nil {
		return t.runFunc(ctx, args, stdio)
	}
	return nil
}

func TestExecute_NoArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	stdio := Stdio{Out: &out, Err: &errOut}
	cmds := []Command{&testCommand{name: "serve", desc: "start server"}}

	code := Execute([]string{}, stdio, cmds)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("expected usage text, got %q", out.String())
	}
}

func TestExecute_Help(t *testing.T) {
	var out bytes.Buffer
	stdio := Stdio{Out: &out}
	cmds := []Command{&testCommand{name: "serve", desc: "start server"}}

	code := Execute([]string{"help"}, stdio, cmds)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "serve") {
		t.Errorf("expected command list, got %q", out.String())
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	var errOut bytes.Buffer
	stdio := Stdio{Err: &errOut}
	cmds := []Command{&testCommand{name: "serve", desc: "start server"}}

	code := Execute([]string{"unknown"}, stdio, cmds)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "unknown command") {
		t.Errorf("expected unknown command error, got %q", errOut.String())
	}
}

func TestExecute_CommandError(t *testing.T) {
	var errOut bytes.Buffer
	stdio := Stdio{Err: &errOut}
	cmds := []Command{&testCommand{
		name: "fail",
		runFunc: func(_ context.Context, _ []string, _ Stdio) error {
			return errors.New("boom")
		},
	}}

	code := Execute([]string{"fail"}, stdio, cmds)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "boom") {
		t.Errorf("expected error message, got %q", errOut.String())
	}
}

func TestExecute_Success(t *testing.T) {
	var out bytes.Buffer
	stdio := Stdio{Out: &out}
	cmds := []Command{&testCommand{
		name: "echo",
		runFunc: func(_ context.Context, args []string, s Stdio) error {
			s.Println("hello", len(args))
			return nil
		},
	}}

	code := Execute([]string{"echo", "a", "b"}, stdio, cmds)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "hello 2") {
		t.Errorf("expected output, got %q", out.String())
	}
}
