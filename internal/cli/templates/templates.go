package templates

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed module/*.tmpl
var moduleFS embed.FS

// ModuleData holds template values for module generation.
type ModuleData struct {
	Name       string // original module name, e.g. "products"
	Package    string // snake_case package name, same as Name
	Struct     string // PascalCase module name, e.g. "Products"
	Plural     string // PascalCase plural, e.g. "Products"
	Table      string // snake_case plural table name, e.g. "products"
	CRUD       bool
	Migration  bool
	Tenant     bool
}

// ParseModuleTemplate loads and parses a named template from the embedded FS.
func ParseModuleTemplate(name string) (*template.Template, error) {
	content, err := moduleFS.ReadFile("module/" + name + ".tmpl")
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", name, err)
	}
	return template.New(name).Parse(string(content))
}

// FileNames returns the list of template file names (without .tmpl extension).
func FileNames() []string {
	return []string{
		"model",
		"dto",
		"repository",
		"service",
		"handler",
		"routes",
		"validation",
	}
}

// NormalizePackage ensures the module name is valid snake_case.
func NormalizePackage(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// StructName returns the PascalCase module name without singularization.
func StructName(name string) string {
	name = NormalizePackage(name)
	if name == "" {
		return ""
	}
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

// PluralStructName returns the PascalCase collection name.
func PluralStructName(name string) string {
	return StructName(name)
}

// TableName returns the snake_case plural table name.
func TableName(name string) string {
	return NormalizePackage(name)
}
