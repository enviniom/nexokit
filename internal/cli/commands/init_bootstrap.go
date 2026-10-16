package commands

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type bootstrapTarget struct {
	Host, Name, User, SSLMode string
	Port                      int
}

// localConfigTarget validates syntax only; it never resolves driver defaults,
// service files, certificates, credentials or database connectivity.
func localConfigTarget(values map[string]string) (bootstrapTarget, error) {
	invalid := errors.New("invalid effective configuration; no files written")
	for _, key := range []string{"APP_PORT", "DB_PORT", "SHUTDOWN_TIMEOUT_SECONDS", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME_SECONDS", "LOG_MAX_SIZE", "LOG_MAX_BACKUPS", "LOG_MAX_AGE", "ACCESS_TTL_MINUTES", "REFRESH_TTL_DAYS", "REDIS_PORT", "REDIS_DB", "REDIS_DIAL_TIMEOUT_SECONDS", "RATE_LIMIT_GLOBAL_RPM", "RATE_LIMIT_LOGIN_RPM", "RATE_LIMIT_REFRESH_RPM", "RATE_LIMIT_WINDOW_SECONDS", "RATE_LIMIT_CLEANUP_INTERVAL_MINUTES"} {
		if value := values[key]; value != "" {
			if _, err := strconv.Atoi(value); err != nil {
				return bootstrapTarget{}, invalid
			}
		}
	}
	target := bootstrapTarget{Host: values["DB_HOST"], Name: values["DB_NAME"], User: values["DB_USER"], SSLMode: values["DB_SSL_MODE"]}
	target.Port, _ = strconv.Atoi(values["DB_PORT"])
	if raw := values["DATABASE_URL"]; raw != "" {
		uri, err := url.Parse(raw)
		if err != nil || (uri.Scheme != "postgres" && uri.Scheme != "postgresql") || uri.Hostname() == "" || uri.User == nil || uri.User.Username() == "" || strings.Trim(uri.Path, "/") == "" {
			return bootstrapTarget{}, invalid
		}
		query, err := url.ParseQuery(uri.RawQuery)
		if err != nil {
			return bootstrapTarget{}, invalid
		}
		for _, key := range []string{"passfile", "service", "servicefile"} {
			if _, ok := query[key]; ok {
				return bootstrapTarget{}, invalid
			}
		}
		target = bootstrapTarget{Host: uri.Hostname(), Name: strings.TrimPrefix(uri.Path, "/"), User: uri.User.Username(), Port: 5432, SSLMode: "prefer"}
		if uri.Port() != "" {
			target.Port, err = strconv.Atoi(uri.Port())
			if err != nil {
				return bootstrapTarget{}, invalid
			}
		}
		// These URL parameters override the corresponding authority components.
		for key, dest := range map[string]*string{"host": &target.Host, "dbname": &target.Name, "user": &target.User, "sslmode": &target.SSLMode} {
			if query.Has(key) {
				*dest = query.Get(key)
			}
		}
		if query.Has("port") {
			target.Port, err = strconv.Atoi(query.Get("port"))
			if err != nil {
				return bootstrapTarget{}, invalid
			}
		}
	}
	if target.Host == "" || target.Name == "" || target.User == "" || target.Port < 1 || target.Port > 65535 || !validInitSSL(target.SSLMode) {
		return bootstrapTarget{}, invalid
	}
	return target, nil
}
