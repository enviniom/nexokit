package projectinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupEffectiveEnvironment(t *testing.T) {
	for _, rebuild := range []bool{false, true} {
		t.Run(map[bool]string{false: "keep", true: "rebuild"}[rebuild], func(t *testing.T) {
			root := fixture(t)
			put(t, root, ".env.example", "APP_NAME=nexokit\n# invented template\nUNKNOWN=template\nDB_SSL_MODE=disable\n")
			put(t, root, ".env", "UNKNOWN=kept\nDB_HOST=kept.invalid\n")
			t.Setenv("UNKNOWN", "ambient")
			p, err := New(root, "Local App", "example.com/local/app")
			if err != nil {
				t.Fatal(err)
			}
			if rebuild {
				if err := p.ConfigureEnvironment(map[string]string{"DB_HOST": "selected.invalid"}); err != nil {
					t.Fatal(err)
				}
			}
			values, err := p.EffectiveEnvironment()
			if err != nil {
				t.Fatal("effective environment unavailable")
			}
			wantUnknown, wantHost := "kept", "kept.invalid"
			if rebuild {
				wantUnknown, wantHost = "template", "selected.invalid"
				if values["APP_ENV"] != "development" || values["DATABASE_URL"] != "" || values["DB_SSL_MODE"] != "disable" {
					t.Error("generated settings/defaults lost")
				}
			}
			if values["UNKNOWN"] != wantUnknown || values["DB_HOST"] != wantHost {
				t.Error("wrong effective environment")
			}
			values["UNKNOWN"] = "mutated caller map"
			again, err := p.EffectiveEnvironment()
			if err != nil || again["UNKNOWN"] != wantUnknown {
				t.Error("effective snapshot exposed mutable state")
			}
			data, err := os.ReadFile(filepath.Join(root, ".env"))
			if err != nil || string(data) != "UNKNOWN=kept\nDB_HOST=kept.invalid\n" {
				t.Error("planning mutated kept environment")
			}
		})
	}
}

func TestSetupKeptSnapshotChecksBytesModeAndIdentity(t *testing.T) {
	for _, change := range []string{"bytes", "mode", "identity"} {
		t.Run(change, func(t *testing.T) {
			directory := fixture(t)
			put(t, directory, ".env", "UNKNOWN=kept\n")
			p, err := New(directory, "App", "example.com/app")
			if err != nil {
				t.Fatal(err)
			}
			selected, err := p.SelectEnvironment()
			if err != nil {
				t.Fatal(err)
			}
			if err := selected.Check(); err != nil {
				t.Fatal("unchanged snapshot refused")
			}
			path := filepath.Join(directory, ".env")
			switch change {
			case "bytes":
				if err := os.WriteFile(path, []byte("UNKNOWN=kept\n# changed\n"), 0644); err != nil {
					t.Fatal(err)
				}
			case "mode":
				if err := os.Chmod(path, 0600); err != nil {
					t.Fatal(err)
				}
			case "identity":
				if err := os.Rename(path, filepath.Join(directory, "old-synthetic-env")); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("UNKNOWN=kept\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := selected.Check(); err == nil {
				t.Error("changed snapshot accepted")
			}
		})
	}
}
