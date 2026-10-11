package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/enviniom/nexokit/internal/cli/templates"
	"github.com/enviniom/nexokit/internal/infra/db"
)

// ModuleOptions controls what the generator produces.
type ModuleOptions struct {
	Name      string
	CRUD      bool
	Migration bool
	Tenant    bool
}

// ErrModuleExists is returned when the target module directory already exists.
var ErrModuleExists = fmt.Errorf("module directory already exists")

var validModuleName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidateName checks that a module name is valid snake_case.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}
	if !validModuleName.MatchString(name) {
		return fmt.Errorf("module name must be snake_case starting with a lowercase letter (got %q)", name)
	}
	return nil
}

// GenerateModule creates the module directory and files.
// If the target directory already exists, it returns ErrModuleExists.
func GenerateModule(opts ModuleOptions) error {
	opts.Name = strings.TrimSpace(opts.Name)
	if err := ValidateName(opts.Name); err != nil {
		return err
	}

	moduleDir := filepath.Join("internal", "modules", opts.Name)
	if _, err := os.Stat(moduleDir); err == nil {
		return fmt.Errorf("%w: %s", ErrModuleExists, moduleDir)
	}

	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}
	createdModuleDir := true
	defer func() {
		if createdModuleDir {
			_ = os.RemoveAll(moduleDir)
		}
	}()

	data := templates.ModuleData{
		Name:      opts.Name,
		Package:   templates.NormalizePackage(opts.Name),
		Struct:    templates.StructName(opts.Name),
		Plural:    templates.PluralStructName(opts.Name),
		Table:     templates.TableName(opts.Name),
		CRUD:      opts.CRUD,
		Migration: opts.Migration,
		Tenant:    opts.Tenant,
	}

	for _, file := range templates.ModuleFiles(data) {
		tmpl, err := templates.ParseModuleTemplate(file.Template)
		if err != nil {
			return err
		}
		fileData := data
		fileData.Action, fileData.UseCase = file.Action, file.UseCase
		if err := executeTemplate(tmpl, moduleDir, file.Path, fileData); err != nil {
			return err
		}
	}

	if opts.Migration {
		if err := db.CreateMigration("migrations", opts.Name); err != nil {
			return fmt.Errorf("failed to create migration: %w", err)
		}
	}

	createdModuleDir = false
	return nil
}

func executeTemplate(tmpl *template.Template, dir, filename string, data templates.ModuleData) error {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filename, err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", filename, err)
	}
	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format %s: %w", filename, err)
	}
	if err := os.WriteFile(path, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}
	return nil
}
