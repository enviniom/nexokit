package projectinit

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
)

var environmentKeys = strings.Fields("APP_NAME APP_ENV APP_URL APP_PORT DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD DB_SSL_MODE DATABASE_URL CORS_ALLOWED_ORIGINS PASETO_KEY")
var assignment = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*[=:]`)

// EnvironmentExists checks metadata only; existing regular files are never read.
func (p *Plan) EnvironmentExists() (bool, error) {
	if err := safePath(p.root); err != nil {
		return false, errors.New("environment ancestor refused")
	}
	info, err := os.Lstat(filepath.Join(p.root, ".env"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil || !info.Mode().IsRegular() {
		return false, errors.New("environment destination refused")
	}
	return true, nil
}

// AddEnvironment prepares optional creation from the identity-transformed snapshot.
// Existing regular destinations are skipped; values are never exposed in diagnostics.
func (p *Plan) AddEnvironment(values map[string]string) error {
	exists, err := p.EnvironmentExists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	known := map[string]bool{}
	for _, k := range environmentKeys {
		known[k] = true
	}
	expected := map[string]string{}
	for k, v := range values {
		if !known[k] || k == "APP_NAME" {
			return errors.New("unsupported environment override")
		}
		for _, r := range v {
			if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
				return errors.New("unsafe environment value")
			}
		}
		expected[k] = v
	}
	expected["APP_ENV"] = "development"
	expected["DATABASE_URL"] = ""
	var template string
	for _, f := range p.files {
		if f.path == ".env.example" {
			template = string(f.after)
		}
	}
	parsed, err := godotenv.Unmarshal(template)
	if err != nil {
		return errors.New("invalid environment template")
	}
	expected["APP_NAME"] = parsed["APP_NAME"]
	lines := strings.SplitAfter(template, "\n")
	seen := map[string]bool{}
	for i, line := range lines {
		match := assignment.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		k := match[1]
		if !known[k] {
			continue
		}
		if seen[k] {
			return errors.New("duplicate environment template key")
		}
		seen[k] = true
		// A target assignment must be independently parseable: reject multiline values.
		one, e := godotenv.Unmarshal(line)
		if e != nil || len(one) != 1 {
			return errors.New("unsupported environment template assignment")
		}
		if v, ok := expected[k]; ok {
			ending := ""
			if strings.HasSuffix(line, "\r\n") {
				ending = "\r\n"
			} else if strings.HasSuffix(line, "\n") {
				ending = "\n"
			}
			lines[i] = k + "=" + encodeEnvironment(v) + ending
		}
	}
	result := strings.Join(lines, "")
	ending := "\n"
	if strings.Contains(template, "\r\n") {
		ending = "\r\n"
	}
	keys := []string{}
	for k := range expected {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if result != "" && !strings.HasSuffix(result, "\n") {
			result += ending
		}
		result += k + "=" + encodeEnvironment(expected[k]) + ending
	}
	verified, e := godotenv.Unmarshal(result)
	if e != nil {
		return errors.New("invalid generated environment")
	}
	for k, v := range expected {
		if verified[k] != v {
			return errors.New("environment value cannot round trip safely")
		}
	}
	if p.environment == nil {
		p.Changes = append(p.Changes, Change{Path: ".env"})
		sort.Slice(p.Changes, func(i, j int) bool { return p.Changes[i].Path < p.Changes[j].Path })
	}
	p.environment = []byte(result)
	return nil
}

func encodeEnvironment(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, `$`, `\$`)
	return `"` + v + `"`
}
