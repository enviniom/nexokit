package cli

import "fmt"

// Printf writes a formatted string to Out.
func (s Stdio) Printf(format string, a ...any) (int, error) {
	return fmt.Fprintf(s.Out, format, a...)
}

// Println writes a line to Out.
func (s Stdio) Println(a ...any) (int, error) {
	return fmt.Fprintln(s.Out, a...)
}

// Errorf writes a formatted string to Err.
func (s Stdio) Errorf(format string, a ...any) (int, error) {
	return fmt.Fprintf(s.Err, format, a...)
}
