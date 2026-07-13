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

func TestEntityMappersHaveUnarySignatures(t *testing.T) {
	file := parseGoFile(t, filepath.Join(queryPackageDir(t), "map_errors.go"))
	expected := map[string]bool{"MapUserError": false, "MapRefreshTokenError": false}

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "MapNotFound" {
			t.Fatal("generic MapNotFound mapper must not remain")
		}
		if _, ok := expected[fn.Name.Name]; !ok {
			continue
		}
		if !hasUnaryErrorSignature(fn) {
			t.Fatalf("%s must accept exactly one error and return error", fn.Name.Name)
		}
		expected[fn.Name.Name] = true
	}

	for name, found := range expected {
		if !found {
			t.Errorf("missing %s", name)
		}
	}
}

func TestRepositoriesUseUnaryEntityMappers(t *testing.T) {
	root := filepath.Dir(queryPackageDir(t))
	paths, err := authRepositoryFiles(root)
	if err != nil || len(paths) == 0 {
		t.Fatalf("discover auth repositories: %v", err)
	}
	for _, path := range paths {
		validateRepositoryFile(t, path)
	}
}

func TestAuthRepositoryFilesDiscoverNestedRepositories(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "slices", "future", "nested", "repository.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create nested fixture: %v", err)
	}
	if err := os.WriteFile(path, []byte("package future\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	paths, err := authRepositoryFiles(root)
	if err != nil {
		t.Fatalf("discover nested repositories: %v", err)
	}
	if len(paths) != 1 || paths[0] != path {
		t.Fatalf("expected nested repository discovery, got %v", paths)
	}
}

func TestRepositoryBoundaryGuardRejectsRawExposure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "repository.go")
	source := `package example
import "github.com/enviniom/nexokit/internal/platform/apperror"
type Repository interface { Save() *apperror.AppError }
func Save(db interface{ Create(any) interface{ Error() error } }) error { return db.Create(nil).Error() }`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	problems := repositoryBoundaryProblems(path)
	if len(problems) < 2 {
		t.Fatalf("guard accepted raw apperror exposure and direct .Error return: %v", problems)
	}
}

func TestRepositoryBoundaryGuardRejectsNestedRawErrorVariable(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "slices", "future", "nested", "repository.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create nested fixture: %v", err)
	}
	source := `package example
import "gorm.io/gorm"
func Save(db *gorm.DB) error {
	result := db.Create(nil)
	persistenceErr := result.Error
	return persistenceErr
}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	problems := repositoryBoundaryProblems(path)
	if len(problems) == 0 {
		t.Fatal("guard accepted a raw .Error stored in a variable and returned later")
	}
}

func validateRepositoryFile(t *testing.T, path string) {
	t.Helper()
	if problems := repositoryBoundaryProblems(path); len(problems) > 0 {
		t.Errorf("%s violates repository boundary: %s", path, strings.Join(problems, "; "))
	}
}

func repositoryBoundaryProblems(path string) []string {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return []string{err.Error()}
	}

	var problems []string
	parents := astParents(file)
	queryAliases := queryImportAliases(file)
	rawVariables := map[string]struct{}{}
	mappedVariables := map[string]struct{}{}
	rawSelectors := map[string]struct{}{}
	mappedSelectors := map[string]struct{}{}
	for _, declaration := range file.Decls {
		switch node := declaration.(type) {
		case *ast.GenDecl:
			if node.Tok == token.IMPORT {
				for _, spec := range node.Specs {
					if strings.Contains(strings.Trim(spec.(*ast.ImportSpec).Path.Value, `"`), "/apperror") {
						problems = append(problems, "imports apperror")
					}
				}
			}
		case *ast.FuncDecl:
			if node.Body == nil {
				continue
			}
			ast.Inspect(node.Body, func(child ast.Node) bool {
				switch statement := child.(type) {
				case *ast.AssignStmt:
					for index, value := range statement.Rhs {
						if isEntityMapperCall(value, queryAliases) {
							markAssignedNames(mappedVariables, statement.Lhs, index)
							continue
						}
						if isPersistenceError(value, parents) || isRawVariable(value, rawVariables) || isPotentialErrorResult(value, statement.Lhs, index) {
							markAssignedNames(rawVariables, statement.Lhs, index)
						}
					}
				case *ast.CallExpr:
					if !isEntityMapperCall(statement, queryAliases) {
						return true
					}
					for _, argument := range statement.Args {
						markMappedError(argument, rawVariables, mappedSelectors)
					}
				case *ast.SelectorExpr:
					if !isPersistenceError(statement, parents) {
						return true
					}
					rawSelectors[errorSelectorKey(statement)] = struct{}{}
				}
				return true
			})

			ast.Inspect(node.Body, func(child ast.Node) bool {
				returnStmt, ok := child.(*ast.ReturnStmt)
				if !ok {
					return true
				}
				for _, result := range returnStmt.Results {
					if isEntityMapperCall(result, queryAliases) || isMappedVariable(result, mappedVariables) {
						continue
					}
					if isPersistenceError(result, parents) || isRawVariable(result, rawVariables) {
						problems = append(problems, node.Name.Name+" returns an unmapped persistence error")
					}
				}
				return true
			})
		}
	}
	for selector := range rawSelectors {
		if _, mapped := mappedSelectors[selector]; !mapped {
			problems = append(problems, "GORM .Error does not pass through an entity mapper: "+selector)
		}
	}

	ast.Inspect(file, func(node ast.Node) bool {
		interfaceType, ok := node.(*ast.InterfaceType)
		if !ok || interfaceType.Methods == nil {
			return true
		}
		for _, field := range interfaceType.Methods.List {
			function, ok := field.Type.(*ast.FuncType)
			if !ok || function.Results == nil {
				continue
			}
			for _, result := range function.Results.List {
				if isAppErrorType(result.Type) {
					problems = append(problems, "repository interface exposes apperror")
				}
			}
		}
		return true
	})
	return problems
}

func authRepositoryFiles(authRoot string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(filepath.Join(authRoot, "slices"), func(path string, entry os.DirEntry, err error) error {
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

func astParents(file *ast.File) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) > 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})
	return parents
}

func queryImportAliases(file *ast.File) map[string]struct{} {
	aliases := map[string]struct{}{}
	for _, imported := range file.Imports {
		if strings.Trim(imported.Path.Value, "\"") != "github.com/enviniom/nexokit/internal/modules/auth/queries" {
			continue
		}
		alias := "queries"
		if imported.Name != nil {
			alias = imported.Name.Name
		}
		aliases[alias] = struct{}{}
	}
	return aliases
}

func isEntityMapperCall(expression ast.Expr, aliases map[string]struct{}) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || (selector.Sel.Name != "MapUserError" && selector.Sel.Name != "MapRefreshTokenError") {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	_, imported := aliases[packageName.Name]
	return imported
}

func isPersistenceError(expression ast.Expr, parents map[ast.Node]ast.Node) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Error" {
		return false
	}
	parent, methodCall := parents[selector].(*ast.CallExpr)
	return !methodCall || parent.Fun != selector
}

func errorSelectorKey(selector *ast.SelectorExpr) string {
	if identifier, ok := selector.X.(*ast.Ident); ok {
		return identifier.Name + ".Error"
	}
	return "<expression>.Error"
}

func markAssignedNames(target map[string]struct{}, lhs []ast.Expr, rhsIndex int) {
	if len(lhs) == 0 {
		return
	}
	index := rhsIndex
	if len(lhs) > 1 {
		index = len(lhs) - 1
	}
	identifier, ok := lhs[index].(*ast.Ident)
	if ok {
		target[identifier.Name] = struct{}{}
	}
}

func isRawVariable(expression ast.Expr, variables map[string]struct{}) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return false
	}
	_, raw := variables[identifier.Name]
	return raw
}

func isMappedVariable(expression ast.Expr, variables map[string]struct{}) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return false
	}
	_, mapped := variables[identifier.Name]
	return mapped
}

func isPotentialErrorResult(expression ast.Expr, lhs []ast.Expr, _ int) bool {
	if _, call := expression.(*ast.CallExpr); !call || len(lhs) < 2 {
		return false
	}
	identifier, ok := lhs[len(lhs)-1].(*ast.Ident)
	return ok && (identifier.Name == "err" || strings.HasSuffix(identifier.Name, "Err"))
}

func markMappedError(expression ast.Expr, rawVariables, mappedSelectors map[string]struct{}) {
	if selector, ok := expression.(*ast.SelectorExpr); ok && selector.Sel.Name == "Error" {
		mappedSelectors[errorSelectorKey(selector)] = struct{}{}
	}
	if identifier, ok := expression.(*ast.Ident); ok {
		delete(rawVariables, identifier.Name)
	}
}

func isAppErrorType(expression ast.Expr) bool {
	star, ok := expression.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == "apperror" && selector.Sel.Name == "AppError"
}

func hasUnaryErrorSignature(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	param, parameterIsError := fn.Type.Params.List[0].Type.(*ast.Ident)
	result, resultIsError := fn.Type.Results.List[0].Type.(*ast.Ident)
	return parameterIsError && resultIsError && param.Name == "error" && result.Name == "error"
}

func queryPackageDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Dir(file)
}

func parseGoFile(t *testing.T, path string) *ast.File {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}
