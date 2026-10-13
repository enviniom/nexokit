package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
)

func runInteractiveInit(ctx context.Context, stdio cli.Stdio) error {
	if err := ctx.Err(); err != nil {
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
		return strings.TrimSpace(line), err
	}
	name, err := read("Application display name: ")
	if err != nil {
		return fmt.Errorf("application input: %w", err)
	}
	slug, err := repositorySlug(name)
	if err != nil {
		return err
	}
	username, err := read("GitHub username: ")
	if err != nil {
		return fmt.Errorf("GitHub username input: %w", err)
	}
	if !validGitHubUsername(username) {
		return errors.New("invalid GitHub username")
	}
	module := "github.com/" + strings.ToLower(username) + "/" + slug
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	plan, err := projectinit.New(root, name, module)
	if err != nil {
		return err
	}
	stdio.Printf("Generated Go module: %s\n", module)
	printInitPreview(stdio, plan)
	answer, err := read("Apply this plan? [y/N]: ")
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	// EOF is always a cancellation, including an unterminated affirmative answer.
	if errors.Is(err, io.EOF) || answer == "" || strings.EqualFold(answer, "no") || strings.EqualFold(answer, "n") {
		stdio.Println("Canceled. No files written.")
		return nil
	}
	if !strings.EqualFold(answer, "yes") && !strings.EqualFold(answer, "y") {
		return errors.New("invalid confirmation; no files written")
	}
	return applyInit(ctx, stdio, plan)
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
