package commands

import (
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
)

func validateInitSetup(values map[string]string) (bootstrapTarget, error) {
	key, err := hex.DecodeString(values["PASETO_KEY"])
	if err != nil || len(key) != 32 {
		return bootstrapTarget{}, errors.New("invalid PASETO key; no files written")
	}
	target, err := localConfigTarget(values)
	if err != nil {
		return bootstrapTarget{}, errors.New("invalid effective configuration; no files written")
	}
	return target, nil
}

func printInitSetupPreview(stdio cli.Stdio, target bootstrapTarget) {
	stdio.Printf("Database target: host=%s port=%d database=%s user=%s sslmode=%s\n", target.Host, target.Port, target.Name, target.User, target.SSLMode)
	stdio.Println("Local files only; database target not contacted. No migrations, seeds or root creation run.")
}

func executeInitSetup(ctx context.Context, stdio cli.Stdio, plan *projectinit.Plan, selected *projectinit.SelectedEnvironment) error {
	if err := ctx.Err(); err != nil {
		return errors.New("initial setup canceled; no files written")
	}
	if err := selected.Check(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return errors.New("initial setup canceled; no files written")
	}
	if err := selected.Check(); err != nil {
		return err
	}
	if err := plan.Apply(); err != nil {
		return fmt.Errorf("initial setup file publication failed: %w", err)
	}
	stdio.Println("Local setup complete: application identity and environment ready.")
	stdio.Println("Review migrations and seeds before executing them. Then run:")
	stdio.Println("  make migrate-up\n  make seed\n  make create-root\n  make dev")
	return nil
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
