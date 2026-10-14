package projectinit

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureEnvironment(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "create", true: "rebuild"}[existing], func(t *testing.T) {
			root := fixture(t)
			put(t, root, ".env.example", "APP_NAME=nexokit\r\n# invented comment\r\nUNKNOWN=untouched\r\n")
			old := "# invented original\r\nSECRET=fixture-only\r\n"
			if existing {
				put(t, root, ".env", old)
			}
			p, err := New(root, "Local App", "example.com/local/app")
			if err != nil {
				t.Fatal(err)
			}
			password := `a ' " # $HOME $(CMD) \\ x=:`
			if err = p.ConfigureEnvironment(map[string]string{"DB_PASSWORD": password}); err != nil {
				t.Fatal(err)
			}
			backup := ""
			for _, change := range p.Changes {
				if strings.HasPrefix(change.Path, ".env.backup-") {
					backup = change.Path
				}
			}
			if existing && backup == "" {
				t.Fatal("missing backup preview")
			}
			if backup != "" {
				if _, err = os.Lstat(filepath.Join(root, backup)); !os.IsNotExist(err) {
					t.Fatal("planning created backup")
				}
			}
			if existing && get(t, root, ".env") != old {
				t.Fatal("planning changed original")
			}
			if err = p.Apply(); err != nil {
				t.Fatal(err)
			}
			text := get(t, root, ".env")
			if !strings.Contains(text, "# invented comment\r\nUNKNOWN=untouched\r\n") {
				t.Fatal("unknown content lost")
			}
			values, err := godotenv.Unmarshal(text)
			if err != nil || values["DB_PASSWORD"] != password || values["APP_ENV"] != "development" || values["DATABASE_URL"] != "" {
				t.Fatal("round trip failed")
			}
			for _, path := range []string{".env", backup} {
				if path == "" {
					continue
				}
				info, err := os.Stat(filepath.Join(root, path))
				if err != nil || info.Mode().Perm() != 0600 {
					t.Fatal("unprotected file")
				}
			}
			if existing && get(t, root, backup) != old {
				t.Fatal("backup differs")
			}
		})
	}
}

func TestConfigureEnvironmentRejections(t *testing.T) {
	for _, scenario := range []string{"collision", "changed", "symlink", "publish", "race", "rollback", "cleanup"} {
		t.Run(scenario, func(t *testing.T) {
			root := fixture(t)
			old := "invented original bytes\r\n"
			put(t, root, ".env", old)
			if err := os.Chmod(filepath.Join(root, ".env"), 0640); err != nil {
				t.Fatal(err)
			}
			originalModule := get(t, root, "go.mod")
			p, err := New(root, "Local App", "example.com/local/app")
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "symlink" {
				// Replacing an existing destination with a link is never accepted.
				if err = os.Rename(filepath.Join(root, ".env"), filepath.Join(root, "saved")); err != nil {
					t.Fatal(err)
				}
				if err = os.Symlink(filepath.Join(root, "saved"), filepath.Join(root, ".env")); err != nil {
					t.Fatal(err)
				}
				if err = p.ConfigureEnvironment(nil); err == nil {
					t.Fatal("accepted symlink")
				}
				return
			}
			if err = p.ConfigureEnvironment(nil); err != nil {
				t.Fatal(err)
			}
			backup := ""
			for _, c := range p.Changes {
				if strings.HasPrefix(c.Path, ".env.backup-") {
					backup = c.Path
				}
			}
			if backup == "" {
				t.Fatal("missing backup preview")
			}
			switch scenario {
			case "collision":
				put(t, root, backup, "competitor")
			case "changed":
				put(t, root, ".env", "competitor")
			case "race":
				p.publishEnvironment = func(from, to string) error {
					if err := os.WriteFile(to, []byte("competitor"), 0600); err != nil {
						return err
					}
					return os.Link(from, to)
				}
			case "publish":
				p.publishEnvironment = func(string, string) error { return errors.New("invented sensitive failure") }
			case "cleanup":
				p.removeTemporary = func(string) error { return errors.New("invented sensitive failure") }
			}
			// Force identity publication failure without bypassing Apply.
			if scenario == "rollback" {
				p.publishEnvironment = func(from, to string) error {
					if err := os.Link(from, to); err != nil {
						return err
					}
					if err := os.Rename(filepath.Join(root, "README.md"), filepath.Join(root, "moved-readme")); err != nil {
						return err
					}
					return os.Symlink(filepath.Join(root, "moved-readme"), filepath.Join(root, "README.md"))
				}
			}
			err = p.Apply()
			if err == nil || strings.Contains(err.Error(), "invented sensitive failure") || strings.Contains(err.Error(), old) {
				t.Fatal("missing or sensitive rejection")
			}
			if scenario == "cleanup" {
				if get(t, root, ".env") == old {
					t.Fatal("cleanup undid publication")
				}
				return
			}
			if scenario == "race" {
				if get(t, root, ".env") != "competitor" || get(t, root, backup) != old {
					t.Fatal("race recovery lost")
				}
				return
			}
			if get(t, root, "go.mod") != originalModule {
				t.Fatal("identity changed on rejection")
			}
			want := old
			if scenario == "changed" {
				want = "competitor"
			}
			if get(t, root, ".env") != want {
				t.Fatal("original/competitor not preserved")
			}
			info, err := os.Stat(filepath.Join(root, ".env"))
			if err != nil || info.Mode().Perm() != 0640 {
				t.Fatal("original mode not restored")
			}
			if scenario == "collision" && get(t, root, backup) != "competitor" {
				t.Fatal("backup overwritten")
			}
		})
	}
}

func TestEnvironmentCreation(t *testing.T) {
	root := fixture(t)
	p, err := New(root, "Local App", "example.com/local/app")
	if err != nil {
		t.Fatal(err)
	}
	if err = p.AddEnvironment(map[string]string{"DB_PASSWORD": `a ' " # $HOME $(CMD) \\ x=:`, "APP_ENV": "production", "DATABASE_URL": "conflict"}); err != nil {
		t.Fatal(err)
	}
	if err = p.Apply(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	values, err := godotenv.Unmarshal(string(b))
	if err != nil {
		t.Fatal("invalid generated environment")
	}
	if values["APP_NAME"] != "Local App" || values["APP_ENV"] != "development" || values["DATABASE_URL"] != "" || values["DB_PASSWORD"] != `a ' " # $HOME $(CMD) \\ x=:` {
		t.Fatal("round trip mismatch")
	}
	info, _ := os.Stat(filepath.Join(root, ".env"))
	if info.Mode().Perm() != 0600 {
		t.Fatal("unsafe mode")
	}
}

func TestEnvironmentSafety(t *testing.T) {
	for _, scenario := range []string{"existing", "symlink", "directory", "race", "publish", "identity", "changed", "missing", "duplicate", "malformed", "unknown", "injection"} {
		t.Run(scenario, func(t *testing.T) {
			root := fixture(t)
			original := get(t, root, "go.mod")
			if scenario == "duplicate" {
				put(t, root, ".env.example", "APP_NAME=nexokit\nDB_PASSWORD=secret-fixture\nDB_PASSWORD=other\n")
			}
			if scenario == "malformed" {
				put(t, root, ".env.example", "APP_NAME=nexokit\nDB_PASSWORD=\"secret-fixture\n")
			}
			p, e := New(root, "Local App", "example.com/local/app")
			if e != nil {
				t.Fatal(e)
			}
			destination := filepath.Join(root, ".env")
			switch scenario {
			case "existing":
				put(t, root, ".env", "secret-fixture")
				exists, e := p.EnvironmentExists()
				if e != nil || !exists {
					t.Fatal("metadata skip")
				}
			case "symlink":
				if e = os.Symlink(filepath.Join(root, "outside"), destination); e != nil {
					t.Fatal(e)
				}
			case "directory":
				if e = os.Mkdir(destination, 0700); e != nil {
					t.Fatal(e)
				}
			}
			values := map[string]string{"DB_PASSWORD": "secret-fixture"}
			if scenario == "unknown" {
				values["UNKNOWN"] = "secret-fixture"
			}
			if scenario == "injection" {
				values["DB_PASSWORD"] = "secret-fixture\nINJECT=x"
			}
			e = p.AddEnvironment(values)
			reject := scenario == "symlink" || scenario == "directory" || scenario == "duplicate" || scenario == "malformed" || scenario == "unknown" || scenario == "injection"
			if reject {
				if e == nil {
					t.Fatal("accepted unsafe input")
				}
				if strings.Contains(e.Error(), "secret-fixture") {
					t.Fatal("secret diagnostic")
				}
				if get(t, root, "go.mod") != original {
					t.Fatal("planning wrote")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			switch scenario {
			case "race":
				p.publishEnvironment = func(from, to string) error {
					info, e := os.Stat(from)
					if e != nil || info.Mode().Perm() != 0600 {
						t.Fatal("unsafe stage")
					}
					put(t, root, ".env", "raced")
					return os.Link(from, to)
				}
			case "publish":
				p.publishEnvironment = func(string, string) error { return errors.New("secret-fixture") }
			case "changed":
				put(t, root, "README.md", "changed")
			case "missing":
				if e = os.Remove(filepath.Join(root, "README.md")); e != nil {
					t.Fatal(e)
				}
			}
			if scenario == "identity" {
				e = p.apply(func(string, string) error { return errors.New("identity failure") })
			} else {
				e = p.Apply()
			}
			if scenario == "existing" {
				if e != nil || get(t, root, ".env") != "secret-fixture" {
					t.Fatal("existing overwritten")
				}
				return
			}
			if e == nil {
				t.Fatal("failure expected")
			}
			if strings.Contains(e.Error(), "secret-fixture") {
				t.Fatal("secret diagnostic")
			}
			if get(t, root, "go.mod") != original {
				t.Fatal("partial identity")
			}
			if scenario == "race" {
				if get(t, root, ".env") != "raced" {
					t.Fatal("race overwritten")
				}
			} else if _, e = os.Lstat(destination); !os.IsNotExist(e) {
				t.Fatal("environment left behind")
			}
			temps, e := filepath.Glob(filepath.Join(root, ".nexokit-init-*"))
			if e != nil || len(temps) != 0 {
				t.Fatal("temporary files left")
			}
		})
	}
}

func TestEnvironmentPreservesRawTemplate(t *testing.T) {
	root := fixture(t)
	put(t, root, ".env.example", "# invented comment\r\nAPP_NAME=nexokit\r\nUNKNOWN=raw # keep\r\nDB_PASSWORD=old\r\nAPP_ENV=production\r\nDATABASE_URL=conflict\r\n")
	p, e := New(root, "Local App", "example.com/local/app")
	if e != nil {
		t.Fatal(e)
	}
	if e = p.AddEnvironment(map[string]string{"DB_PASSWORD": `\n \" '$HOME ${HOME} $(CMD) # :=`}); e != nil {
		t.Fatal(e)
	}
	if e = p.Apply(); e != nil {
		t.Fatal(e)
	}
	got := get(t, root, ".env")
	if !strings.Contains(got, "# invented comment\r\n") || !strings.Contains(got, "UNKNOWN=raw # keep\r\n") {
		t.Fatal("raw content lost")
	}
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Fatal("CRLF lost")
	}
}

func TestEnvironmentRollbackOwnership(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "restore", true: "ownership changed"}[replace], func(t *testing.T) {
			root := fixture(t)
			original := get(t, root, "README.md")
			p, e := New(root, "Local App", "example.com/local/app")
			if e != nil {
				t.Fatal(e)
			}
			if e = p.AddEnvironment(map[string]string{"DB_PASSWORD": "secret-fixture"}); e != nil {
				t.Fatal(e)
			}
			calls := 0
			e = p.apply(func(a, b string) error {
				calls++
				if calls == 3 {
					if replace {
						if err := os.Rename(filepath.Join(root, ".env"), filepath.Join(root, "owned-env")); err != nil {
							t.Fatal(err)
						}
						put(t, root, ".env", "raced")
					}
					return errors.New("identity failure")
				}
				return os.Rename(a, b)
			})
			if e == nil || strings.Contains(e.Error(), "secret-fixture") {
				t.Fatal("unsafe failure")
			}
			if get(t, root, "README.md") != original {
				t.Fatal("identity not restored")
			}
			if replace {
				if !strings.Contains(e.Error(), "ROLLBACK FAILED .env") || get(t, root, ".env") != "raced" {
					t.Fatal("ownership recovery evidence missing")
				}
			} else if _, err := os.Lstat(filepath.Join(root, ".env")); !os.IsNotExist(err) {
				t.Fatal("owned environment remains")
			}
		})
	}
}

func TestEnvironmentStageFailure(t *testing.T) {
	root := fixture(t)
	original := get(t, root, "go.mod")
	p, e := New(root, "Local App", "example.com/local/app")
	if e != nil {
		t.Fatal(e)
	}
	if e = p.AddEnvironment(map[string]string{"DB_PASSWORD": "secret-fixture"}); e != nil {
		t.Fatal(e)
	}
	calls := 0
	p.createTemporary = func(dir, pattern string) (*os.File, error) {
		calls++
		if calls == len(p.Changes)*2-1 {
			return nil, errors.New("injected staging failure")
		}
		return os.CreateTemp(dir, pattern)
	}
	if e = p.Apply(); e == nil {
		t.Fatal("accepted stage failure")
	}
	if get(t, root, "go.mod") != original {
		t.Fatal("identity mutated before staging complete")
	}
	if _, e = os.Lstat(filepath.Join(root, ".env")); !os.IsNotExist(e) {
		t.Fatal("environment created")
	}
	temps, _ := filepath.Glob(filepath.Join(root, ".nexokit-init-*"))
	if len(temps) != 0 {
		t.Fatal("stage leak")
	}
}

func TestEnvironmentCleanupFailureProtectedResidue(t *testing.T) {
	root := fixture(t)
	p, err := New(root, "Local App", "example.com/local/app")
	if err != nil {
		t.Fatal(err)
	}
	if err = p.AddEnvironment(map[string]string{"DB_PASSWORD": "invented-private-cleanup"}); err != nil {
		t.Fatal(err)
	}
	residue := ""
	p.removeTemporary = func(path string) error {
		b, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(b), "invented-private-cleanup") {
			residue = path
			return errors.New("invented-private-cleanup")
		}
		return os.Remove(path)
	}
	err = p.Apply()
	if err == nil || residue == "" || !strings.Contains(err.Error(), residue) || strings.Contains(err.Error(), "invented-private-cleanup") {
		t.Fatal("cleanup failure not redacted/recoverable")
	}
	info, e := os.Stat(residue)
	if e != nil || info.Mode().Perm() != 0600 {
		t.Fatal("residue not protected")
	}
	// Cleanup failure occurs after successful identity/environment publication;
	// never delete an adopter destination as a side effect of temporary cleanup.
	if _, e = os.Stat(filepath.Join(root, ".env")); e != nil {
		t.Fatal("cleanup deleted published environment")
	}
	p.removeTemporary = nil
}
