package commands

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
	"github.com/enviniom/nexokit/internal/cli/root"
)

func runInteractiveInit(ctx context.Context, stdio cli.Stdio) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := projectinit.CheckUninitialized(directory); err != nil {
		return err
	}
	if stdio.In == nil {
		return errors.New("interactive init requires input")
	}
	reader := bufio.NewReader(stdio.In)
	read := func(prompt string) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		stdio.Printf("%s", prompt)
		line, err := reader.ReadString('\n')
		if canceled := ctx.Err(); canceled != nil {
			return "", canceled
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if !safeLocalValue(line) {
			return "", errors.New("invalid input")
		}
		return strings.TrimSpace(line), err
	}
	name, err := read("Application display name: ")
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("application input failed; no files written")
	}
	slug, err := repositorySlug(name)
	if err != nil {
		return err
	}
	username, err := read("GitHub username: ")
	if err != nil {
		return errors.New("GitHub username input failed; no files written")
	}
	if !validGitHubUsername(username) {
		return errors.New("invalid GitHub username")
	}
	module := "github.com/" + strings.ToLower(username) + "/" + slug
	if err := ctx.Err(); err != nil {
		return err
	}
	plan, err := projectinit.New(directory, name, module)
	if err != nil {
		return err
	}
	exists, err := plan.EnvironmentExists()
	if err != nil {
		return err
	}
	rebuild := !exists
	if exists {
		choice, err := read("Existing .env: KEEP or BACKUP+REBUILD? [keep]: ")
		if err != nil {
			return errors.New("environment choice input failed; no files written")
		}
		switch strings.ToLower(choice) {
		case "", "keep":
			stdio.Println("Keeping existing .env bytes and mode; using its exact configuration.")
		case "rebuild", "backup+rebuild":
			rebuild = true
		default:
			return errors.New("invalid environment choice; no files written")
		}
	}
	if rebuild {
		values, err := collectLocalEnvironment(stdio, slug, read, func() (string, error) {
			return readInitPassword(ctx, stdio, reader)
		}, initEntropy)
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := plan.ConfigureEnvironment(values); err != nil {
			return err
		}
	}
	selected, err := plan.SelectEnvironment()
	if err != nil {
		return err
	}
	rootName, err := read("Root name: ")
	if err != nil {
		return errors.New("root input failed; no files written")
	}
	rootEmail, err := read("Root email: ")
	if err != nil {
		return errors.New("root input failed; no files written")
	}
	rootPassword, err := readInitRootPassword(ctx, stdio, reader)
	if err != nil {
		return err
	}
	input := root.CreateRootInput{Name: rootName, Email: rootEmail, Password: rootPassword}
	target, err := validateInitSetup(selected.Values, input)
	if err != nil {
		return err
	}
	stdio.Printf("Generated Go module: %s\n", module)
	printInitPreview(stdio, plan)
	printInitSetupPreview(stdio, target)
	answer, err := read("Apply this plan? [y/N]: ")
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("confirmation input failed; no files written")
	}
	// EOF is always a cancellation, including an unterminated affirmative answer.
	if errors.Is(err, io.EOF) || answer == "" || strings.EqualFold(answer, "no") || strings.EqualFold(answer, "n") {
		stdio.Println("Canceled. No files written.")
		return nil
	}
	if !strings.EqualFold(answer, "yes") && !strings.EqualFold(answer, "y") {
		return errors.New("invalid confirmation; no files written")
	}
	return executeInitSetup(ctx, stdio, plan, selected, input)
}

func validGitHubUsername(username string) bool {
	if len(username) < 1 || len(username) > 39 || username[0] == '-' || username[len(username)-1] == '-' || strings.Contains(username, "--") {
		return false
	}
	for _, c := range username {
		if !asciiAlphanumeric(c) && c != '-' {
			return false
		}
	}
	return true
}

func asciiAlphanumeric(c rune) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

func repositorySlug(name string) (string, error) {
	var slug strings.Builder
	separator := false
	for _, c := range name {
		if c == ' ' || c == '.' || c == '_' || c == '-' {
			separator = slug.Len() > 0
			continue
		}
		if !asciiAlphanumeric(c) {
			return "", errors.New("application name cannot form a safe ASCII repository slug")
		}
		if separator {
			slug.WriteByte('-')
			separator = false
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		slug.WriteRune(c)
	}
	if slug.Len() == 0 {
		return "", errors.New("application name requires a nonempty repository slug")
	}
	return slug.String(), nil
}
