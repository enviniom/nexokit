package queries

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompanyMappersHaveUnarySignatures(t *testing.T) {
	file := parseCompanyFile(t, filepath.Join(companyQueryDir(t), "map_errors.go"))
	expected := map[string]bool{"MapCompanyError": false, "MapCompanyDomainError": false}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if _, wanted := expected[fn.Name.Name]; wanted {
			expected[fn.Name.Name] = unaryErrorSignature(fn)
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("%s must accept exactly one error and return error", name)
		}
	}
}

func TestCompanyRepositoriesUseEntityMappers(t *testing.T) {
	root := filepath.Dir(companyQueryDir(t))
	paths, err := companyRepositoryFiles(root)
	if err != nil || len(paths) != 7 {
		t.Fatalf("discover seven repositories: paths=%v err=%v", paths, err)
	}
	for _, path := range paths {
		if problems := companyRepositoryProblems(path); len(problems) > 0 {
			t.Errorf("%s: %s", path, strings.Join(problems, "; "))
		}
	}
}

func TestCompanyRepositoryGuardFixtures(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		wantProblem  bool
	}{
		{"direct error", `package x; type result struct { Error error }; func f(result result) error { return result.Error }`, true},
		{"variable held", `package x; type result struct { Error error }; func f(result result) error { persistenceErr := result.Error; return persistenceErr }`, true},
		{"nested block", `package x; type result struct { Error error }; func f(result result) error { if true { return result.Error }; return nil }`, true},
		{"single result transaction", `package x; func f(db interface{ Transaction(func(any) error) error }) error { err := db.Transaction(nil); return err }`, true},
		{"same selector separate functions", `package x; import q "github.com/enviniom/nexokit/internal/modules/companies/queries"; type result struct { Error error }; func good(result result) error { return q.MapCompanyError(result.Error) }; func bad(result result) error { return result.Error }`, true},
		{"same function mapped selector then raw selector", `package x; import q "github.com/enviniom/nexokit/internal/modules/companies/queries"; type result struct { Error error }; func f(result result) error { mapped := q.MapCompanyError(result.Error); if mapped != nil { return mapped }; return result.Error }`, true},
		{"mapped variable", `package x; import q "github.com/enviniom/nexokit/internal/modules/companies/queries"; type result struct { Error error }; func f(result result) error { persistenceErr := result.Error; return q.MapCompanyError(persistenceErr) }`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "repository.go")
			if err := os.WriteFile(path, []byte(tt.source), 0o600); err != nil {
				t.Fatal(err)
			}
			got := len(companyRepositoryProblems(path)) > 0
			if got != tt.wantProblem {
				t.Fatalf("problems=%v, want problem=%t", companyRepositoryProblems(path), tt.wantProblem)
			}
		})
	}
}

func companyRepositoryFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(filepath.Join(root, "slices"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == "repository.go" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func companyRepositoryProblems(path string) []string {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return []string{err.Error()}
	}
	var problems []string
	aliases := companyQueryAliases(file)
	for _, imported := range file.Imports {
		if strings.Contains(strings.Trim(imported.Path.Value, `"`), "/apperror") {
			problems = append(problems, "imports apperror")
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		iface, ok := node.(*ast.InterfaceType)
		if !ok {
			return true
		}
		for _, method := range iface.Methods.List {
			if fn, ok := method.Type.(*ast.FuncType); ok && fn.Results != nil {
				for _, result := range fn.Results.List {
					if isCompanyAppError(result.Type) {
						problems = append(problems, "repository interface exposes apperror")
					}
				}
			}
		}
		return true
	})
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		raw := map[string]token.Pos{}
		mapped := map[token.Pos]bool{}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.AssignStmt:
				for i, value := range n.Rhs {
					if isCompanyMapperCall(value, aliases) {
						markCompanyMapped(n.Lhs, i, raw)
						continue
					}
					if isCompanyPersistenceSelector(value) || isCompanyRaw(value, raw) || isCompanySingleErrorResult(value, n.Lhs) {
						markCompanyRaw(n.Lhs, i, raw, companyRawIdentity(value, raw))
					}
				}
			case *ast.CallExpr:
				if isCompanyMapperCall(n, aliases) {
					for _, argument := range n.Args {
						if identity := companyRawIdentity(argument, raw); identity.IsValid() {
							mapped[identity] = true
						}
						if ident, ok := argument.(*ast.Ident); ok {
							delete(raw, ident.Name)
						}
					}
				}
			case *ast.ReturnStmt:
				for _, result := range n.Results {
					if isCompanyMapperCall(result, aliases) {
						continue
					}
					for _, identity := range companyReturnedRawIdentities(result, raw) {
						if !mapped[identity] {
							problems = append(problems, fn.Name.Name+" returns unmapped persistence error")
						}
					}
				}
			}
			return true
		})
	}
	return problems
}

func companyQueryAliases(file *ast.File) map[string]bool {
	aliases := map[string]bool{}
	for _, imported := range file.Imports {
		if strings.Trim(imported.Path.Value, `"`) == "github.com/enviniom/nexokit/internal/modules/companies/queries" {
			name := "queries"
			if imported.Name != nil {
				name = imported.Name.Name
			}
			aliases[name] = true
		}
	}
	return aliases
}
func isCompanyMapperCall(expr ast.Expr, aliases map[string]bool) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || (sel.Sel.Name != "MapCompanyError" && sel.Sel.Name != "MapCompanyDomainError") {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && aliases[pkg.Name]
}
func isCompanyPersistenceSelector(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Error"
}
func companyRawIdentity(expr ast.Expr, raw map[string]token.Pos) token.Pos {
	if selector, ok := expr.(*ast.SelectorExpr); ok && isCompanyPersistenceSelector(selector) {
		return selector.Pos()
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return raw[ident.Name]
	}
	return expr.Pos()
}
func companyReturnedRawIdentities(expr ast.Expr, raw map[string]token.Pos) []token.Pos {
	identities := map[token.Pos]bool{}
	ast.Inspect(expr, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.SelectorExpr:
			if isCompanyPersistenceSelector(n) {
				identities[n.Pos()] = true
			}
		case *ast.Ident:
			if identity := raw[n.Name]; identity.IsValid() {
				identities[identity] = true
			}
		}
		return true
	})
	result := make([]token.Pos, 0, len(identities))
	for identity := range identities {
		result = append(result, identity)
	}
	return result
}
func markCompanyRaw(lhs []ast.Expr, index int, raw map[string]token.Pos, identity token.Pos) {
	if len(lhs) > 1 {
		index = len(lhs) - 1
	}
	if index < len(lhs) {
		if ident, ok := lhs[index].(*ast.Ident); ok {
			raw[ident.Name] = identity
		}
	}
}
func markCompanyMapped(lhs []ast.Expr, index int, raw map[string]token.Pos) {
	if len(lhs) > 1 {
		index = len(lhs) - 1
	}
	if index < len(lhs) {
		if ident, ok := lhs[index].(*ast.Ident); ok {
			delete(raw, ident.Name)
		}
	}
}
func isCompanyRaw(expr ast.Expr, raw map[string]token.Pos) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident == nil {
		return false
	}
	_, found := raw[ident.Name]
	return found
}
func isCompanySingleErrorResult(expr ast.Expr, lhs []ast.Expr) bool {
	_, call := expr.(*ast.CallExpr)
	if !call || len(lhs) == 0 {
		return false
	}
	ident, ok := lhs[len(lhs)-1].(*ast.Ident)
	return ok && (ident.Name == "err" || strings.HasSuffix(ident.Name, "Err"))
}
func isCompanyAppError(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "AppError" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "apperror"
}
func unaryErrorSignature(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Results == nil || len(fn.Type.Params.List) != 1 || len(fn.Type.Results.List) != 1 {
		return false
	}
	p, pOK := fn.Type.Params.List[0].Type.(*ast.Ident)
	r, rOK := fn.Type.Results.List[0].Type.(*ast.Ident)
	return pOK && rOK && p.Name == "error" && r.Name == "error"
}
func companyQueryDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	return filepath.Dir(file)
}
func parseCompanyFile(t *testing.T, path string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	return file
}
