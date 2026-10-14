package commands

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
	"github.com/enviniom/nexokit/internal/cli/root"
	"golang.org/x/term"
)

// Private seam for public-command tests; no production environment bypass.
var initSetupPrepare = prepareInitBootstrap

func validateInitSetup(values map[string]string, input root.CreateRootInput) (bootstrapTarget, error) {
	if err := root.ValidateInput(input); err != nil {
		return bootstrapTarget{}, errors.New("invalid root credentials; no files written")
	}
	key, err := hex.DecodeString(values["PASETO_KEY"])
	if err != nil || len(key) != 32 {
		return bootstrapTarget{}, errors.New("invalid PASETO key; no files written")
	}
	_, _, target, err := bootstrapConfig(values)
	if err != nil {
		return bootstrapTarget{}, errors.New("invalid effective configuration; no files written")
	}
	return target, nil
}

func printInitSetupPreview(stdio cli.Stdio, target bootstrapTarget) {
	stdio.Printf("Database target: host=%s port=%d database=%s user=%s sslmode=%s\n", target.Host, target.Port, target.Name, target.User, target.SSLMode)
	stdio.Println("Setup order: migrations -> seeds -> root")
	stdio.Println("No services provisioned; no transaction spans files and database stages.")
}

func executeInitSetup(ctx context.Context, stdio cli.Stdio, plan *projectinit.Plan, selected *projectinit.SelectedEnvironment, input root.CreateRootInput) error {
	if err := ctx.Err(); err != nil {
		return errors.New("initial setup canceled; no files written")
	}
	if err := selected.Check(); err != nil {
		return err
	}
	prepared, err := initSetupPrepare(ctx, selected.Values, input)
	if err != nil {
		return errors.New("initial setup preparation failed; no files written")
	}
	defer prepared.Close()
	if err := ctx.Err(); err != nil {
		return errors.New("initial setup canceled; no files written")
	}
	if err := selected.Check(); err != nil {
		return err
	}
	if err := plan.Apply(); err != nil {
		return fmt.Errorf("initial setup file publication failed; database stages not run: %w", err)
	}
	if err := prepared.Run(ctx, stdio); err != nil {
		stdio.Println("Identity adopted: init cannot rerun. Recover with migrate, seed and create-root maintenance commands. Published environment backups remain recovery artifacts; no automatic database rollback.")
		return err // Run exposes only its fixed stage name, not driver diagnostics.
	}
	if err := prepared.Close(); err != nil {
		stdio.Println("Identity adopted: init cannot rerun; use maintenance commands for recovery.")
		return errors.New("initial setup connection close failed")
	}
	stdio.Println("Setup complete: application adopted, migrations and seeds applied, root access initialized. cmd/nexokit remains the developer tool.")
	return nil
}

func initRootFromEnvironment() root.CreateRootInput {
	return root.CreateRootInput{Name: os.Getenv("ROOT_USER_NAME"), Email: os.Getenv("ROOT_USER_EMAIL"), Password: os.Getenv("ROOT_USER_PASSWORD")}
}

// Flag mode has no wizard or inherited URL override. Require explicit selected
// component/application values; an intentionally empty DB password is allowed.
func initFlagEnvironment() (map[string]string, error) {
	values := map[string]string{}
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "APP_PORT", "APP_URL", "CORS_ALLOWED_ORIGINS"} {
		value, ok := os.LookupEnv(key)
		if !ok || (!safeLocalValue(value)) || (key != "DB_PASSWORD" && value == "") {
			return nil, errors.New("explicit local application/database configuration required; no files written")
		}
		values[key] = value
	}
	if !localHost(values["DB_HOST"]) || !localPort(values["DB_PORT"]) || !localPort(values["APP_PORT"]) || !localURL(values["APP_URL"], false) {
		return nil, errors.New("invalid local configuration; no files written")
	}
	for _, origin := range strings.Split(values["CORS_ALLOWED_ORIGINS"], ",") {
		if !localURL(strings.TrimSpace(origin), true) {
			return nil, errors.New("invalid local configuration; no files written")
		}
	}
	if value, ok := os.LookupEnv("DB_SSL_MODE"); ok {
		if !validInitSSL(value) {
			return nil, errors.New("invalid local configuration; no files written")
		}
		values["DB_SSL_MODE"] = value
	}
	var key [32]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return nil, errors.New("development key generation failed; no files written")
	}
	values["PASETO_KEY"] = hex.EncodeToString(key[:])
	return values, nil
}

func validInitSSL(value string) bool {
	switch value {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return true
	}
	return false
}

func readInitRootPassword(ctx context.Context, stdio cli.Stdio, reader *bufio.Reader) (string, error) {
	fail := errors.New("root password input failed; no files written")
	if ctx.Err() != nil {
		return "", fail
	}
	stdio.Printf("Root password (hidden): ")
	var value string
	if file, ok := stdio.In.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		if reader.Buffered() != 0 {
			return "", fail
		}
		data, err := term.ReadPassword(int(file.Fd()))
		stdio.Println("")
		if err != nil {
			return "", fail
		}
		value = string(data)
	} else {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fail
		}
		value = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	}
	if ctx.Err() != nil || !safeLocalValue(value) {
		return "", fail
	}
	return value, nil
}
