package commands

import (
	"strings"
	"testing"
)

func TestInitSetupLocalConfigurationSyntax(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		valid     bool
	}{
		{"components", "", true},
		{"url", "postgres://chosen:invented@selected.invalid:5441/owned?sslmode=disable", true},
		{"certificate paths are not read", "postgres://chosen@selected.invalid/owned?sslmode=verify-full&sslrootcert=/nonexistent/synthetic-cert&sslcert=/nonexistent/synthetic-client&sslkey=/nonexistent/synthetic-key", true},
		{"invalid scheme", "file://chosen@selected.invalid/owned", false},
		{"invalid port", "postgres://chosen@selected.invalid:99999/owned", false},
		{"invalid ssl", "postgres://chosen@selected.invalid/owned?sslmode=invalid", false},
		{"service", "postgres://chosen@selected.invalid/owned?service=invented", false},
		{"bad query", "postgres://chosen@selected.invalid/owned?port=%zz", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := selectedBootstrapValues()
			values["DATABASE_URL"] = tc.raw
			t.Setenv("PGSERVICE", "ignored-inherited-service")
			t.Setenv("PGSSLROOTCERT", "/nonexistent/inherited-cert")
			target, err := localConfigTarget(values)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if tc.valid && target.Host != "selected.invalid" {
				t.Fatal("inherited target selected")
			}
			if err != nil && strings.Contains(err.Error(), "invented") {
				t.Fatal("secret leaked")
			}
		})
	}
}
