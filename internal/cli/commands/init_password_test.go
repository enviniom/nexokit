package commands

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

func TestPasswordPipePreservesWhitespace(t *testing.T) {
	for _, input := range []string{"  invented  \n", "  invented  \r\n", "\n"} {
		var output bytes.Buffer
		in := strings.NewReader(input)
		got, err := readInitPassword(context.Background(), cli.Stdio{In: in, Out: &output}, bufio.NewReader(in))
		want := strings.TrimSuffix(strings.TrimSuffix(input, "\n"), "\r")
		if err != nil || got != want {
			t.Fatal("password input changed")
		}
		if want != "" && strings.Contains(output.String(), want) {
			t.Fatal("password echoed")
		}
	}
	for _, input := range []string{"", "unterminated"} {
		in := strings.NewReader(input)
		if _, err := readInitPassword(context.Background(), cli.Stdio{In: in, Out: &bytes.Buffer{}}, bufio.NewReader(in)); err == nil {
			t.Fatal("incomplete password accepted")
		}
	}
}
