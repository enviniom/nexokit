package commands

import (
	"bufio"
	"context"
	"errors"
	"os"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"golang.org/x/term"
)

// Cancellation is checked around synchronous reads; it cannot interrupt a blocked
// descriptor read. No helper goroutine is left behind on cancellation.
func readInitPassword(ctx context.Context, stdio cli.Stdio, reader *bufio.Reader) (string, error) {
	fail := errors.New("password input failed; no files written")
	if ctx.Err() != nil {
		return "", fail
	}
	stdio.Printf("DB_PASSWORD (Enter for intentionally empty local password): ")
	var value string
	if file, ok := stdio.In.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		// A prefetched password must never be discarded or echoed through fallback.
		if reader.Buffered() != 0 {
			return "", fail
		}
		b, err := term.ReadPassword(int(file.Fd()))
		stdio.Println("")
		if err != nil {
			return "", fail
		}
		value = string(b)
	} else {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fail
		}
		value = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	}
	if ctx.Err() != nil {
		return "", fail
	}
	return value, nil
}
