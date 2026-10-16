package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

func TestInitLocalApplyWithoutRootGoOrDatabase(t *testing.T) {
	initFixture(t)
	initSetupProcessSettings(t)
	for _, key := range []string{"ROOT_USER_NAME", "ROOT_USER_EMAIL", "ROOT_USER_PASSWORD"} {
		t.Setenv(key, "")
	}
	t.Setenv("PATH", "")
	t.Setenv("DB_SSL_MODE", "disable")
	t.Setenv("DB_HOST", "unreachable.invalid")
	if err := os.Mkdir("migrations", 0700); err != nil {
		t.Fatal(err)
	}
	const migration = "-- Edit this initial migration before applying it.\n"
	if err := os.WriteFile("migrations/001_initial.sql", []byte(migration), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	exit := cli.Execute([]string{"init", "--name", "App", "--module", "example.com/app", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
	if exit != 0 {
		t.Fatalf("local init exit %d, stderr %q", exit, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr %q", stderr.String())
	}
	module, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(module), "module example.com/app") {
		t.Error("local initialization did not adopt identity")
	}
	after, err := os.ReadFile("migrations/001_initial.sql")
	if err != nil || string(after) != migration {
		t.Error("initial migration changed")
	}
	if _, err := os.Stat(".tmp"); !os.IsNotExist(err) {
		t.Error("init created a seed runner")
	}
	for _, message := range []string{"make migrate-up", "make seed", "make create-root"} {
		if !strings.Contains(out.String(), message) {
			t.Errorf("missing explicit next step %q", message)
		}
	}
	for _, message := range []string{"Running migrations", "Running seeds", "Running root", "root access initialized"} {
		if strings.Contains(out.String(), message) {
			t.Errorf("init claimed database activity: %q", message)
		}
	}
}

func TestInitLocalInteractiveDoesNotAskForRoot(t *testing.T) {
	initFixture(t)
	t.Setenv("PATH", "")
	const kept = "DB_HOST=unreachable.invalid\nDB_PORT=5439\nDB_NAME=local\nDB_USER=local\nDB_PASSWORD=InventedSecret9\nDB_SSL_MODE=disable\nPASETO_KEY=abababababababababababababababababababababababababababababababab\n"
	if err := os.WriteFile(".env", []byte(kept), 0640); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	exit := cli.Execute([]string{"init"}, cli.Stdio{In: strings.NewReader("App\nuser\nkeep\ny\n"), Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
	if exit != 0 || stderr.Len() != 0 {
		t.Fatalf("local interactive exit %d stderr %q", exit, stderr.String())
	}
	for _, text := range []string{"Root name:", "Root email:", "Root password", "InventedSecret9"} {
		if strings.Contains(out.String()+stderr.String(), text) {
			t.Errorf("unexpected prompt/secret %q", text)
		}
	}
	if strings.Count(out.String(), "Apply this plan?") != 1 {
		t.Error("expected one final confirmation")
	}
	data, err := os.ReadFile(".env")
	if err != nil || string(data) != kept {
		t.Error("kept environment changed")
	}
	info, err := os.Stat(".env")
	if err != nil || info.Mode().Perm() != 0640 {
		t.Error("kept environment mode changed")
	}
}
