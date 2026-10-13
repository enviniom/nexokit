package commands

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/enviniom/nexokit/internal/cli"
)

func safeLocalValue(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return false
		}
	}
	return true
}

func localPort(value string) bool {
	n, err := strconv.Atoi(value)
	return err == nil && n >= 1 && n <= 65535
}

func localURL(value string, origin bool) bool {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || !localHost(u.Hostname()) || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return false
	}
	if u.Port() != "" && !localPort(u.Port()) {
		return false
	}
	return !origin || (u.Path == "" && u.RawPath == "")
}

func localHost(value string) bool {
	if net.ParseIP(value) != nil {
		return true
	}
	for _, r := range value {
		if !asciiAlphanumeric(r) && r != '.' && r != '-' {
			return false
		}
	}
	return value != ""
}

// collectLocalEnvironment never reads defaults from template contents.
func collectLocalEnvironment(stdio cli.Stdio, slug string, read func(string) (string, error), password func() (string, error), entropy io.Reader) (map[string]string, error) {
	values := map[string]string{}
	ask := func(key, fallback string, valid func(string) bool) error {
		value, err := read(key + " [" + fallback + "]: ")
		if err != nil {
			return errors.New("local configuration input failed; no files written")
		}
		if value == "" {
			value = fallback
		}
		if !safeLocalValue(value) || !valid(value) {
			return errors.New("invalid local configuration; no files written")
		}
		values[key] = value
		return nil
	}
	plain := func(v string) bool { return v != "" && safeLocalValue(v) }
	for _, field := range []struct {
		key, fallback string
		valid         func(string) bool
	}{
		{"DB_HOST", "localhost", localHost}, {"DB_PORT", "5432", localPort}, {"DB_NAME", slug, plain}, {"DB_USER", "postgres", plain},
	} {
		if err := ask(field.key, field.fallback, field.valid); err != nil {
			return nil, err
		}
	}
	secret, err := password()
	if err != nil || !safeLocalValue(secret) {
		return nil, errors.New("password input failed; no files written")
	}
	values["DB_PASSWORD"] = secret
	ssl := func(v string) bool {
		switch v {
		case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
			return true
		}
		return false
	}
	if err := ask("DB_SSL_MODE", "disable", ssl); err != nil {
		return nil, err
	}
	if err := ask("APP_PORT", "8080", localPort); err != nil {
		return nil, err
	}
	if err := ask("APP_URL", "http://localhost:"+values["APP_PORT"], func(v string) bool { return localURL(v, false) }); err != nil {
		return nil, err
	}
	if err := ask("CORS_ALLOWED_ORIGINS", "http://localhost:3000", func(v string) bool {
		for _, origin := range strings.Split(v, ",") {
			if !localURL(strings.TrimSpace(origin), true) {
				return false
			}
		}
		return true
	}); err != nil {
		return nil, err
	}
	var key [32]byte
	if _, err := io.ReadFull(entropy, key[:]); err != nil {
		return nil, errors.New("development key generation failed; no files written")
	}
	values["PASETO_KEY"] = hex.EncodeToString(key[:])
	stdio.Println("Local development only; no database provisioned.")
	for _, field := range []string{"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_SSL_MODE", "APP_PORT", "APP_URL", "CORS_ALLOWED_ORIGINS"} {
		stdio.Printf("%s: %s\n", field, values[field])
	}
	stdio.Println("DB_PASSWORD: [hidden]\nPASETO_KEY: [generated]")
	return values, nil
}

var initEntropy io.Reader = rand.Reader
