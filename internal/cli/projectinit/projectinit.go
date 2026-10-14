// Package projectinit plans one-time local starter adoption without external commands.
package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// OriginalModule is assembled so adoption cannot rewrite the canonical identity.
func OriginalModule() string { return "github.com/" + "enviniom/" + "nexokit" }

type Change struct{ Path string }
type snapshot struct {
	path          string
	before, after []byte
	mode          fs.FileMode
}

// Plan exposes paths, never potentially sensitive file contents.
type Plan struct {
	Changes             []Change
	Remaining           []string
	root                string
	files               []snapshot
	environment         []byte
	environmentOriginal *snapshot
	environmentIdentity os.FileInfo
	environmentBackup   string
	// Test seam for exclusive publication; production always uses os.Link.
	publishEnvironment func(string, string) error
	createTemporary    func(string, string) (*os.File, error)
	removeTemporary    func(string) error
}

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 ._-]{0,99}$`)
var modulePart = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]*$`)
var hostPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`)

func validate(name, module string) error {
	if !namePattern.MatchString(name) || strings.TrimSpace(name) != name || strings.Contains(name, "..") {
		return errors.New("name must be 1–100 plain letters, digits, spaces, dots, hyphens or underscores; no traversal")
	}
	parts := strings.Split(module, "/")
	if len(parts) < 2 || !hostPattern.MatchString(parts[0]) || !strings.Contains(parts[0], ".") || strings.Contains(parts[0], "..") {
		return errors.New("module must be a Go import path with a lowercase domain and application path")
	}
	for _, part := range parts[1:] {
		if !modulePart.MatchString(part) || strings.Contains(part, "..") || strings.HasSuffix(part, ".") {
			return errors.New("unsafe module path component")
		}
	}
	last := parts[len(parts)-1]
	if len(last) > 1 && last[0] == 'v' {
		if n, e := strconv.Atoi(last[1:]); e == nil && (n < 2 || last[1] == '0') {
			return errors.New("module major version suffix must be /v2 or later")
		}
	}
	if module == OriginalModule() || strings.HasPrefix(module, OriginalModule()+"/") {
		return errors.New("choose an independent module path")
	}
	return nil
}

// safePath checks every ancestor, not only the final file. Symlinks are never followed.
func safePath(path string) error {
	for current := path; ; current = filepath.Dir(current) {
		info, e := os.Lstat(current)
		if e != nil {
			return e
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink refused: %s", current)
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return nil
}
func readRegular(path string) ([]byte, fs.FileMode, error) {
	if e := safePath(path); e != nil {
		return nil, 0, e
	}
	info, e := os.Lstat(path)
	if e != nil {
		return nil, 0, e
	}
	if info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
		return nil, 0, fmt.Errorf("special file mode refused: %s", path)
	}
	if !info.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("not a regular file: %s", path)
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, 0, e
	}
	if bytes.ContainsRune(b, 0) {
		return nil, 0, fmt.Errorf("binary content refused: %s", path)
	}
	return b, info.Mode(), nil
}
func excluded(name string) bool {
	return strings.HasPrefix(name, ".") || name == "vendor" || name == "logs" || name == "tmp" || name == "odd" || name == "node_modules" || name == "testdata-cache" || strings.HasPrefix(name, ".env")
}

// replaceModule matches the exact import-path prefix, not similarly named projects.
func replaceModule(text, module string) string {
	old := OriginalModule()
	var out strings.Builder
	for {
		i := strings.Index(text, old)
		if i < 0 {
			out.WriteString(text)
			break
		}
		end := i + len(old)
		valid := end == len(text) || strings.ContainsRune("/\"'` \t\r\n)", rune(text[end]))
		out.WriteString(text[:i])
		if valid {
			out.WriteString(module)
		} else {
			out.WriteString(old)
		}
		text = text[end:]
	}
	return out.String()
}

// New validates identity and reads only known text families. It never writes.
func New(root, name, module string) (*Plan, error) {
	if e := validate(name, module); e != nil {
		return nil, e
	}
	absolute, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	if e = safePath(absolute); e != nil {
		return nil, e
	}
	paths := []string{"go.mod", "README.md", ".env.example"}
	for _, dir := range []string{"cmd", "internal", "tests", "seeds"} {
		base := filepath.Join(absolute, dir)
		if _, e = os.Lstat(base); os.IsNotExist(e) {
			continue
		}
		e = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if excluded(d.Name()) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("source symlink refused: %s", path)
			}
			if d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".tmpl" {
				return nil
			}
			rel, err := filepath.Rel(absolute, path)
			if err != nil {
				return err
			}
			paths = append(paths, filepath.ToSlash(rel))
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	for _, doc := range []string{"docs/starter-template.md", "docs/references/runtime/cli.md"} {
		if _, e := os.Lstat(filepath.Join(absolute, doc)); e == nil {
			paths = append(paths, doc)
		} else if !os.IsNotExist(e) {
			return nil, e
		}
	}
	sort.Strings(paths)
	p := &Plan{root: absolute}
	anchors := map[string]bool{}
	for _, path := range paths {
		before, mode, e := readRegular(filepath.Join(absolute, filepath.FromSlash(path)))
		if e != nil {
			return nil, fmt.Errorf("preflight %s: %w", path, e)
		}
		text := string(before)
		if strings.HasSuffix(path, ".go") {
			parsed, err := parser.ParseFile(token.NewFileSet(), path, before, parser.ImportsOnly)
			if err != nil {
				return nil, fmt.Errorf("parse candidate %s: %w", path, err)
			}
			for _, spec := range parsed.Imports {
				importPath, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return nil, err
				}
				if strings.Contains(importPath, "/internal/") && !strings.HasPrefix(importPath, OriginalModule()+"/internal/") {
					return nil, fmt.Errorf("unexpected nonstarter internal import in %s: %s", path, importPath)
				}
			}
		}
		after := replaceModule(text, module)
		switch path {
		case "go.mod":
			count := 0
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				fields := strings.Fields(line)
				if len(fields) > 0 && fields[0] == "module" {
					count++
					if len(fields) != 2 || fields[1] != OriginalModule() {
						return nil, errors.New("not an uninitialized NexoKit module")
					}
					lines[i] = strings.Replace(line, OriginalModule(), module, 1)
				}
			}
			after = strings.Join(lines, "\n")
			anchors[path] = count == 1
		case "README.md":
			anchors[path] = strings.HasPrefix(text, "# NexoKit\n")
			after = strings.Replace(after, "# NexoKit\n", "# "+name+"\n", 1)
		case ".env.example":
			count := 0
			for _, line := range strings.Split(text, "\n") {
				line = strings.TrimSuffix(line, "\r")
				if strings.HasPrefix(line, "APP_NAME=") {
					count++
					if line != "APP_NAME=nexokit" {
						return nil, errors.New("example APP_NAME is not the starter identity")
					}
				}
			}
			anchors[path] = count == 1
			value := name
			if strings.Contains(name, " ") {
				value = strconv.Quote(name)
			}
			lines := strings.Split(after, "\n")
			for i, line := range lines {
				if strings.TrimSuffix(line, "\r") == "APP_NAME=nexokit" {
					ending := ""
					if strings.HasSuffix(line, "\r") {
						ending = "\r"
					}
					lines[i] = "APP_NAME=" + value + ending
				}
			}
			after = strings.Join(lines, "\n")
		case "cmd/nexokit/main.go":
			anchors[path] = strings.Contains(text, `"`+OriginalModule()+`/internal/cli`)
		}
		if (strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".tmpl")) && strings.Contains(after, OriginalModule()) {
			return nil, fmt.Errorf("unexpected residual starter module in %s", path)
		}
		if strings.Contains(after, "NexoKit") || strings.Contains(after, "nexokit") {
			p.Remaining = append(p.Remaining, path)
		}
		p.files = append(p.files, snapshot{path: path, before: before, after: []byte(after), mode: mode})
		if after != text {
			p.Changes = append(p.Changes, Change{Path: path})
		}
	}
	for _, path := range []string{"go.mod", "README.md", ".env.example", "cmd/nexokit/main.go"} {
		if !anchors[path] {
			return nil, fmt.Errorf("missing starter identity anchor: %s", path)
		}
	}
	return p, nil
}

// Apply preflights all snapshots (even unchanged candidates), stages all replacements,
// then renames. On a commit error it restores completed replacements and reports rollback errors.
// Run on a quiescent local copy: this is not a cross-process filesystem transaction.
func (p *Plan) Apply() error { return p.apply(os.Rename) }
func (p *Plan) apply(rename func(string, string) error) (result error) {
	if p.environment != nil {
		exists, err := p.EnvironmentExists()
		if err != nil {
			return err
		}
		if p.environmentOriginal != nil {
			if !exists || p.checkEnvironmentOriginal() != nil {
				return errors.New("environment changed since preview")
			}
			if _, err := os.Lstat(filepath.Join(p.root, p.environmentBackup)); !os.IsNotExist(err) {
				return errors.New("backup destination refused")
			}
		} else if exists {
			return errors.New("environment destination appeared since preview")
		}
	}
	for _, f := range p.files {
		before, mode, e := readRegular(filepath.Join(p.root, f.path))
		if e != nil {
			return fmt.Errorf("preflight %s: %w", f.path, e)
		}
		if mode != f.mode || !bytes.Equal(before, f.before) {
			return fmt.Errorf("candidate changed since preview: %s", f.path)
		}
	}
	type staged struct {
		file         snapshot
		next, backup string
		owned        os.FileInfo
	}
	var stages []staged
	var temporary []string
	defer func() {
		for _, path := range temporary {
			if path == "" {
				continue
			}
			remove := p.removeTemporary
			if remove == nil {
				remove = os.Remove
			}
			if e := remove(path); e != nil && !os.IsNotExist(e) {
				result = errors.Join(result, fmt.Errorf("cleanup failed: %s; inspect protected residue manually", path))
			}
		}
	}()
	stage := func(f snapshot, data []byte) (string, error) {
		create := p.createTemporary
		if create == nil {
			create = os.CreateTemp
		}
		file, e := create(filepath.Dir(filepath.Join(p.root, f.path)), ".nexokit-init-*")
		if e != nil {
			return "", e
		}
		temporary = append(temporary, file.Name())
		_, writeErr := file.Write(data)
		modeErr := file.Chmod(f.mode.Perm())
		syncErr := file.Sync()
		closeErr := file.Close()
		return file.Name(), errors.Join(writeErr, modeErr, syncErr, closeErr)
	}
	for _, f := range p.files {
		if bytes.Equal(f.before, f.after) {
			continue
		}
		next, e := stage(f, f.after)
		if e != nil {
			return fmt.Errorf("stage %s: %w", f.path, e)
		}
		backup, e := stage(f, f.before)
		if e != nil {
			return fmt.Errorf("backup %s: %w", f.path, e)
		}
		owned, e := os.Lstat(next)
		if e != nil {
			return errors.New("stage metadata failed")
		}
		stages = append(stages, staged{file: f, next: next, backup: backup, owned: owned})
	}
	if p.environment != nil {
		next, err := stage(snapshot{path: ".env", mode: 0600}, p.environment)
		if err != nil {
			return errors.New("stage environment failed")
		}
		owned, err := os.Lstat(next)
		if err != nil {
			return errors.New("environment stage metadata failed")
		}
		if err = safePath(p.root); err != nil {
			return errors.New("environment ancestor refused")
		}
		publish := p.publishEnvironment
		if publish == nil {
			publish = os.Link
		}
		destination := filepath.Join(p.root, ".env")
		recovery := ""
		if p.environmentOriginal != nil {
			// The protected public backup is never overwritten or cleaned automatically.
			recovery, err = stage(snapshot{path: ".env", mode: 0600}, p.environmentOriginal.before)
			if err != nil {
				return errors.New("stage environment backup failed")
			}
			if err = p.checkEnvironmentOriginal(); err != nil {
				return err
			}
			if err = os.Link(recovery, filepath.Join(p.root, p.environmentBackup)); err != nil {
				return errors.New("exclusive backup publication failed")
			}
			// Keep a separate restoration inode with the original permissions.
			recovery, err = stage(snapshot{path: ".env", mode: p.environmentOriginal.mode}, p.environmentOriginal.before)
			if err != nil {
				return errors.New("stage environment recovery failed")
			}
			if err = p.checkEnvironmentOriginal(); err != nil {
				return err
			}
			if err = os.Remove(destination); err != nil {
				return errors.New("environment replacement refused")
			}
		}
		restore := func() error {
			if recovery == "" {
				return nil
			}
			if err := os.Link(recovery, destination); err != nil {
				for k, path := range temporary {
					if path == recovery {
						temporary[k] = ""
					}
				}
				return fmt.Errorf("ROLLBACK FAILED .env: protected recovery copy %s; destination preserved", recovery)
			}
			return nil
		}
		if err = publish(next, destination); err != nil {
			return errors.Join(errors.New("exclusive environment publication failed"), restore())
		}
		defer func() {
			if result == nil {
				return
			}
			current, err := os.Lstat(destination)
			if os.IsNotExist(err) {
				result = errors.Join(result, restore())
				return
			}
			if err != nil || !os.SameFile(owned, current) {
				result = errors.Join(result, errors.New("ROLLBACK FAILED .env: destination ownership changed; inspect .env manually"))
				return
			}
			if err = safePath(p.root); err == nil {
				err = os.Remove(destination)
			}
			if err != nil {
				result = errors.Join(result, errors.New("ROLLBACK FAILED .env: remove transaction-owned .env manually"))
			} else {
				result = errors.Join(result, restore())
			}
		}()
	}
	for i, s := range stages {
		e := safePath(filepath.Join(p.root, s.file.path))
		if e == nil {
			e = rename(s.next, filepath.Join(p.root, s.file.path))
		}
		if e != nil {
			result = fmt.Errorf("apply %s: %w", s.file.path, e)
			for j := i - 1; j >= 0; j-- {
				prev := stages[j]
				destination := filepath.Join(p.root, prev.file.path)
				rollbackErr := safePath(destination)
				if rollbackErr == nil {
					current, err := os.Lstat(destination)
					if err != nil || !os.SameFile(current, prev.owned) {
						rollbackErr = errors.New("destination ownership changed")
					} else {
						rollbackErr = os.Rename(prev.backup, destination)
					}
				}
				if rollbackErr != nil {
					// Preserve the recovery copy when restoration fails, rather than deleting it.
					for k, path := range temporary {
						if path == prev.backup {
							temporary[k] = ""
						}
					}
					result = errors.Join(result, fmt.Errorf("ROLLBACK FAILED %s (recovery copy %s): %w", prev.file.path, prev.backup, rollbackErr))
				}
			}
			return result
		}
	}
	return nil
}
