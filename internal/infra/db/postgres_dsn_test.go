package db

import (
	"github.com/enviniom/nexokit/internal/config"
	"github.com/jackc/pgx/v5"
	"os"
	"testing"
)

func TestDSNLiteralCredentialsAndOverride(t *testing.T) {
	t.Setenv("PGPASSFILE", os.DevNull)
	for _, password := range []string{"simple123", "spaces #$ 'quotes' \\ backslash", ""} {
		t.Run(password, func(t *testing.T) {
			cfg := &config.Config{DB: config.DBConfig{Host: "127.0.0.1", Port: 5439, Name: "owned", User: "fake", Password: password, SSLMode: "disable"}}
			parsed, err := pgx.ParseConfig(buildDSN(cfg))
			if err != nil {
				t.Fatal("DSN did not parse")
			}
			if parsed.Password != password || parsed.Host != "127.0.0.1" || parsed.Port != 5439 || parsed.Database != "owned" || parsed.User != "fake" {
				t.Fatal("DSN changed credentials or target")
			}
			if password == "simple123" && buildDSN(cfg) != "host=127.0.0.1 user=fake password=simple123 dbname=owned port=5439 sslmode=disable" {
				t.Fatal("legacy simple DSN changed")
			}
			cfg.DB.DatabaseURL = "postgres://fake:invented@127.0.0.1:5439/owned?sslmode=disable"
			if buildDSN(cfg) != cfg.DB.DatabaseURL {
				t.Fatal("URL override changed")
			}
		})
	}
}
