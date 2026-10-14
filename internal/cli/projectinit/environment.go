package projectinit

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
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
	return p.prepareEnvironment(values)
}

func (p *Plan) prepareEnvironment(values map[string]string) error {
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

// ConfigureEnvironment plans creation or an explicitly selected backup/rebuild.
// Planning is write-free; Apply publishes the previewed backup path exclusively.
func (p *Plan) ConfigureEnvironment(values map[string]string) error {
	if p.environment != nil {
		return errors.New("environment already planned")
	}
	exists, err := p.EnvironmentExists()
	if err != nil {
		return err
	}
	var original *snapshot
	var identity os.FileInfo
	backup := ""
	if exists {
		path := filepath.Join(p.root, ".env")
		data, mode, e := readRegular(path)
		if e != nil {
			return errors.New("environment snapshot refused")
		}
		identity, e = os.Lstat(path)
		if e != nil {
			return errors.New("environment metadata refused")
		}
		original = &snapshot{path: ".env", before: data, mode: mode}
		var nonce [16]byte
		if _, e = rand.Read(nonce[:]); e != nil {
			return errors.New("backup name unavailable")
		}
		backup = ".env.backup-" + hex.EncodeToString(nonce[:])
		if _, e = os.Lstat(filepath.Join(p.root, backup)); !os.IsNotExist(e) {
			return errors.New("backup destination refused")
		}
	}
	if err = p.prepareEnvironment(values); err != nil {
		return err
	}
	p.environmentOriginal, p.environmentIdentity, p.environmentBackup = original, identity, backup
	if backup != "" {
		p.Changes = append(p.Changes, Change{Path: backup})
		sort.Slice(p.Changes, func(i, j int) bool { return p.Changes[i].Path < p.Changes[j].Path })
	}
	return nil
}

func (p *Plan) checkEnvironmentOriginal() error {
	path := filepath.Join(p.root, ".env")
	data, mode, err := readRegular(path)
	if err != nil {
		return errors.New("environment changed since preview")
	}
	info, err := os.Lstat(path)
	if err != nil || !os.SameFile(info, p.environmentIdentity) || mode != p.environmentOriginal.mode || !bytes.Equal(data, p.environmentOriginal.before) {
		return errors.New("environment changed since preview")
	}
	return nil
}

func encodeEnvironment(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, `$`, `\$`)
	return `"` + v + `"`
}
