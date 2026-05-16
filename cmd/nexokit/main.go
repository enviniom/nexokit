package main

import (
	"os"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/commands"
)

func main() {
	stdio := cli.Stdio{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	os.Exit(cli.Execute(os.Args[1:], stdio, commands.All()))
}
